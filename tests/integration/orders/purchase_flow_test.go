package orders_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/app/routers"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const buyerPassword = "Password123!"

// registerBuyer creates a user whose password can actually be used to log in.
func registerBuyer(t *testing.T, email string) *models.User {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(buyerPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &models.User{Name: "Buyer", Email: email, Password: string(hashed)}
	require.NoError(t, database.DB.Create(user).Error)
	return user
}

// login exchanges the seeded credentials for an access token.
func login(t *testing.T, router *gin.Engine, email string) string {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, buyerPassword)
	rec := do(t, router, http.MethodPost, "/api/v1/auth/login", body, "")
	require.Equal(t, http.StatusOK, rec.Code, "login should succeed: %s", rec.Body.String())

	var envelope struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.NotEmpty(t, envelope.Data.AccessToken)
	return envelope.Data.AccessToken
}

func do(t *testing.T, router *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// purchaseData pulls the purchase payload out of the standard envelope.
func purchaseData(t *testing.T, rec *httptest.ResponseRecorder) struct {
	OrderID    uint   `json:"order_id"`
	TicketID   uint   `json:"ticket_id"`
	TicketCode string `json:"ticket_code"`
	EventID    uint   `json:"event_id"`
	PriceCents int64  `json:"price_cents"`
} {
	t.Helper()
	var envelope struct {
		Data struct {
			OrderID    uint   `json:"order_id"`
			TicketID   uint   `json:"ticket_id"`
			TicketCode string `json:"ticket_code"`
			EventID    uint   `json:"event_id"`
			PriceCents int64  `json:"price_cents"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	return envelope.Data
}

func TestPurchaseEndpoint(t *testing.T) {
	router := routers.SetupRoute()

	t.Run("positive", func(t *testing.T) {
		t.Run("authenticated buyer with stock gets a ticket", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", 3, 15000000)
			registerBuyer(t, "buyer1@example.local")
			token := login(t, router, "buyer1@example.local")

			rec := do(t, router, http.MethodPost,
				fmt.Sprintf("/api/v1/events/%d/purchase", event.ID), "{}", token)

			// Covers AE1
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			got := purchaseData(t, rec)
			assert.NotEmpty(t, got.TicketCode)
			assert.Equal(t, int64(15000000), got.PriceCents)
			assert.Equal(t, int64(1), countTickets(t, event.ID, models.TicketStatusSold))

			var orders int64
			require.NoError(t, database.DB.Model(&models.Order{}).Count(&orders).Error)
			assert.Equal(t, int64(1), orders)
		})

		t.Run("two purchases by the same buyer take different tickets", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", 3, 15000000)
			registerBuyer(t, "buyer2@example.local")
			token := login(t, router, "buyer2@example.local")

			path := fmt.Sprintf("/api/v1/events/%d/purchase", event.ID)
			first := purchaseData(t, do(t, router, http.MethodPost, path, "{}", token))
			second := purchaseData(t, do(t, router, http.MethodPost, path, "{}", token))

			assert.NotEqual(t, first.TicketID, second.TicketID)
			assert.NotEqual(t, first.OrderID, second.OrderID)
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("without a token the request is rejected", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", 3, 15000000)

			rec := do(t, router, http.MethodPost,
				fmt.Sprintf("/api/v1/events/%d/purchase", event.ID), "{}", "")

			// Covers AE3
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("unknown event is rejected", func(t *testing.T) {
			harness.Reset(t)
			registerBuyer(t, "buyer3@example.local")
			token := login(t, router, "buyer3@example.local")

			rec := do(t, router, http.MethodPost, "/api/v1/events/999999/purchase", "{}", token)

			// Covers AE4
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("sold out event is rejected and creates no order", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", 1, 15000000)
			registerBuyer(t, "buyer4@example.local")
			token := login(t, router, "buyer4@example.local")

			path := fmt.Sprintf("/api/v1/events/%d/purchase", event.ID)
			require.Equal(t, http.StatusCreated, do(t, router, http.MethodPost, path, "{}", token).Code)

			rec := do(t, router, http.MethodPost, path, "{}", token)

			// Covers AE2
			assert.Equal(t, http.StatusConflict, rec.Code)
			var orders int64
			require.NoError(t, database.DB.Model(&models.Order{}).Count(&orders).Error)
			assert.Equal(t, int64(1), orders, "the refused purchase left no order behind")
		})

		t.Run("non numeric event id is rejected", func(t *testing.T) {
			harness.Reset(t)
			registerBuyer(t, "buyer5@example.local")
			token := login(t, router, "buyer5@example.local")

			rec := do(t, router, http.MethodPost, "/api/v1/events/abc/purchase", "{}", token)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	})

	t.Run("edge case", func(t *testing.T) {
		t.Run("changing the event price does not rewrite a completed order", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", 2, 15000000)
			registerBuyer(t, "buyer6@example.local")
			token := login(t, router, "buyer6@example.local")

			rec := do(t, router, http.MethodPost,
				fmt.Sprintf("/api/v1/events/%d/purchase", event.ID), "{}", token)
			require.Equal(t, http.StatusCreated, rec.Code)
			got := purchaseData(t, rec)

			require.NoError(t, database.DB.Model(&models.Event{}).
				Where("id = ?", event.ID).Update("price_cents", 99900000).Error)

			// Covers AE5
			var order models.Order
			require.NoError(t, database.DB.First(&order, got.OrderID).Error)
			assert.Equal(t, int64(15000000), order.PriceCents,
				"the order keeps the price the buyer actually paid")
		})
	})
}
