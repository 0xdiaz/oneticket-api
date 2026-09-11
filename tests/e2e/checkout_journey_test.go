// Checkout journeys: buying a ticket over the wire, and every way that is
// refused. Split out of journey_test.go so neither file crosses the 300-line
// limit in docs/00_AI_CRITICAL_RULES.md.
package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
)

type purchase struct {
	ID          uint      `json:"id"`
	EventID     uint      `json:"event_id"`
	TicketID    uint      `json:"ticket_id"`
	TicketCode  string    `json:"ticket_code"`
	PriceCents  int64     `json:"price_cents"`
	PurchasedAt time.Time `json:"purchased_at"`
}

// seedEventWithTickets writes one event and n available tickets straight
// through database.DB.
//
// This is the one place in this package that touches the database directly,
// and it touches it only to build fixtures. No HTTP route creates an event,
// and the harness runs migrations without the dev seeders (those live in
// main.go behind APP_ENV=development), so the database this package starts
// against is empty and nothing is on sale. The rule E2E actually keeps is
// about the request and assertion path, and that stays pure: every claim
// below is made through call, over a real socket.
//
// tests/e2e never calls harness.Reset, so rows survive for the whole package.
// Each journey therefore seeds its own event and test order decides nothing.
func seedEventWithTickets(t *testing.T, name string, tickets int, saleStartsAt time.Time) *models.Event {
	t.Helper()

	ev := &models.Event{
		Name:         name,
		Venue:        "E2E Hall",
		StartsAt:     time.Now().Add(24 * time.Hour),
		SaleStartsAt: saleStartsAt,
		PriceCents:   15000000,
		TotalTickets: tickets,
	}
	if err := database.DB.Create(ev).Error; err != nil {
		t.Fatalf("seed event %q: %v", name, err)
	}

	rows := make([]models.Ticket, 0, tickets)
	for i := 1; i <= tickets; i++ {
		rows = append(rows, models.Ticket{
			EventID: ev.ID,
			Code:    fmt.Sprintf("A-%03d", i),
			Status:  models.TicketStatusAvailable,
		})
	}
	if tickets > 0 {
		if err := database.DB.Create(&rows).Error; err != nil {
			t.Fatalf("seed %d tickets for %q: %v", tickets, name, err)
		}
	}
	return ev
}

// signIn registers a fresh account and returns the access token the real login
// endpoint mints for it. A token a test hands itself would prove nothing.
func signIn(t *testing.T, prefix string) string {
	t.Helper()

	email := uniqueEmail(prefix)
	const password = "correct-horse-battery"

	status, env := call(t, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"name": "E2E Buyer", "email": email, "password": password,
	})
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("register answered %d, want 200 or 201: %s", status, env.Message)
	}

	status, env = call(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email": email, "password": password,
	})
	if status != http.StatusOK {
		t.Fatalf("login answered %d, want 200: %s", status, env.Message)
	}
	var issued tokens
	decode(t, env, &issued)
	if issued.AccessToken == "" {
		t.Fatal("login succeeded but returned an empty access_token")
	}
	return issued.AccessToken
}

// purchasePath takes any id, including the malformed ones, because a client
// can send anything in that segment and the server still owes it an answer.
func purchasePath(id any) string {
	return fmt.Sprintf("/api/v1/events/%v/purchase", id)
}

// readEvent fetches one event over HTTP, which is the only way this layer is
// allowed to observe availability.
func readEvent(t *testing.T, id uint) event {
	t.Helper()

	status, env := call(t, http.MethodGet, fmt.Sprintf("/api/v1/events/%d", id), "", nil)
	if status != http.StatusOK {
		t.Fatalf("GET /events/%d answered %d, want 200: %s", id, status, env.Message)
	}
	var got event
	decode(t, env, &got)
	return got
}

// TestJourney_VisitorBuysATicket is the checkout path end to end: arrive, sign
// up, sign in, see what is on sale, buy, and watch the inventory move.
//
// The last step is the one no unit test can make. Availability is counted from
// the tickets table on read, so a purchase that answered 201 without claiming
// a row would still look like a success to the buyer and show up only here.
func TestJourney_VisitorBuysATicket(t *testing.T) {
	seeded := seedEventWithTickets(t, "E2E Flash Sale", 100, time.Now().Add(-time.Minute))
	token := signIn(t, "buyer")

	status, env := call(t, http.MethodGet, "/api/v1/events", "", nil)
	if status != http.StatusOK {
		t.Fatalf("GET /events answered %d, want 200: %s", status, env.Message)
	}
	var listed []event
	decode(t, env, &listed)
	var onSale bool
	for _, e := range listed {
		if e.ID == seeded.ID {
			onSale = true
			break
		}
	}
	if !onSale {
		t.Fatalf("GET /events did not list the seeded event %d", seeded.ID)
	}
	availableBefore := readEvent(t, seeded.ID).AvailableTickets

	status, env = call(t, http.MethodPost, purchasePath(seeded.ID), token, nil)
	if status != http.StatusCreated {
		t.Fatalf("purchase answered %d, want 201: %s", status, env.Message)
	}
	if !env.Success {
		t.Error("a successful purchase came back with success=false")
	}

	// price_cents has to arrive as a JSON number. Money is int64 at every
	// layer here, and a quoted value on the wire is how that invariant
	// escapes without any Go test noticing.
	var fields map[string]json.RawMessage
	decode(t, env, &fields)
	raw, ok := fields["price_cents"]
	if !ok {
		t.Fatalf("purchase payload has no price_cents field: %s", env.Data)
	}
	if strings.HasPrefix(strings.TrimSpace(string(raw)), `"`) {
		t.Errorf("price_cents came back as a JSON string (%s), want a number", raw)
	}

	var bought purchase
	decode(t, env, &bought)
	if bought.TicketCode == "" {
		t.Error("purchase succeeded but returned an empty ticket_code")
	}
	if bought.EventID != seeded.ID {
		t.Errorf("purchase reports event_id %d, want %d", bought.EventID, seeded.ID)
	}
	if bought.PriceCents != seeded.PriceCents {
		t.Errorf("purchase reports price_cents %d, want %d", bought.PriceCents, seeded.PriceCents)
	}

	if got := readEvent(t, seeded.ID).AvailableTickets; got != availableBefore-1 {
		t.Errorf("available_tickets went from %d to %d after one purchase, want %d",
			availableBefore, got, availableBefore-1)
	}
}

// dbLeakMarkers are strings only the storage layer produces. Any of them in a
// response body means the driver's own words reached a client.
//
// The list stops at database markers on purpose. AuthMiddleware passes err
// into utils.Unauthorized (internal/app/middlewares/auth.go:43), so a 401 body
// legitimately carries JWT library text such as "token is malformed: ...".
// That is known debt, but the middleware is out of scope for this story, so
// asserting its absence here would fail on code this change may not touch.
var dbLeakMarkers = []string{"pgx", "int4", "OID", "SQL", "sql:"}

// assertNoInternalLeak re-encodes the envelope and scans it. call returns the
// decoded envelope rather than the bytes it read, and client_test.go is not
// this story's file; the re-encoded form still carries every field a client
// sees.
func assertNoInternalLeak(t *testing.T, env envelope) {
	t.Helper()

	body, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("re-encode response envelope: %v", err)
	}
	for _, marker := range dbLeakMarkers {
		if strings.Contains(string(body), marker) {
			t.Errorf("failure response leaked internal detail %q: %s", marker, body)
		}
	}
}

// TestJourney_PurchaseRejections covers every way a checkout can be refused.
// The status code is half of it; the other half is that a refusal says enough
// for a client to act on and nothing about the machinery behind it.
func TestJourney_PurchaseRejections(t *testing.T) {
	token := signIn(t, "buyer-rejected")
	soldOut := seedEventWithTickets(t, "E2E Sold Out", 2, time.Now().Add(-time.Minute))
	notOpen := seedEventWithTickets(t, "E2E Belum Dibuka", 5, time.Now().Add(24*time.Hour))

	const missingEventID = 999999
	var soldOutMessage string

	t.Run("negative", func(t *testing.T) {
		t.Run("event tidak ada menjawab 404", func(t *testing.T) {
			// Gin answers an unregistered path with 404 as well, so a bare
			// status check here would pass even with no route at all.
			// Comparing against the NoRoute fallback separates "no such event"
			// from "no such endpoint" without pinning either message.
			_, fallback := call(t, http.MethodPost, "/api/v1/there-is-no-such-thing", token, nil)

			status, env := call(t, http.MethodPost, purchasePath(missingEventID), token, nil)
			if status != http.StatusNotFound {
				t.Fatalf("purchase of an unknown event answered %d, want 404: %s", status, env.Message)
			}
			if env.Message == fallback.Message {
				t.Errorf("purchase of an unknown event answered with the NoRoute fallback %q; the route is not registered",
					env.Message)
			}
		})

		t.Run("stok habis menjawab 409", func(t *testing.T) {
			for i := 1; i <= 2; i++ {
				status, env := call(t, http.MethodPost, purchasePath(soldOut.ID), token, nil)
				if status != http.StatusCreated {
					t.Fatalf("purchase %d of 2 answered %d, want 201: %s", i, status, env.Message)
				}
			}
			status, env := call(t, http.MethodPost, purchasePath(soldOut.ID), token, nil)
			if status != http.StatusConflict {
				t.Fatalf("purchase against a sold out event answered %d, want 409: %s", status, env.Message)
			}
			soldOutMessage = env.Message
		})

		t.Run("masa jual belum dibuka menjawab 409 dengan pesan berbeda dari stok habis", func(t *testing.T) {
			status, env := call(t, http.MethodPost, purchasePath(notOpen.ID), token, nil)
			if status != http.StatusConflict {
				t.Fatalf("purchase before the sale opens answered %d, want 409: %s", status, env.Message)
			}
			// Pinning the exact wording would turn a copy edit into a failure.
			// What has to hold is that the two 409s are distinguishable: one is
			// worth retrying later, the other never is.
			if soldOutMessage == "" {
				t.Skip("the sold out message was never captured, nothing to compare against")
			}
			if env.Message == soldOutMessage {
				t.Errorf("sale-not-open and sold-out both answer 409 with %q; a client cannot tell them apart",
					env.Message)
			}
		})

		t.Run("id bukan angka menjawab 400", func(t *testing.T) {
			status, env := call(t, http.MethodPost, purchasePath("abc"), token, nil)
			if status != http.StatusBadRequest {
				t.Fatalf("purchase with a non-numeric id answered %d, want 400: %s", status, env.Message)
			}
		})

		t.Run("id di luar jangkauan int4 menjawab 400", func(t *testing.T) {
			// The read path answers 500 to this and leaks the driver's message
			// while doing it. A number too large for the column is a bad
			// request, not a server fault, and the write path must not inherit
			// the mistake.
			status, env := call(t, http.MethodPost, purchasePath("99999999999999"), token, nil)
			if status != http.StatusBadRequest {
				t.Fatalf("purchase with an out-of-range id answered %d, want 400: %s", status, env.Message)
			}
		})

		t.Run("respons gagal tidak membawa detail internal", func(t *testing.T) {
			// Every refusal above, plus the two unauthenticated purchase rows
			// from TestJourney_GuardRejectsUnauthenticated, re-sent here so one
			// place covers all of them.
			cases := []struct {
				name  string
				path  string
				token string
			}{
				{"tanpa token", purchasePath(soldOut.ID), ""},
				{"token sampah", purchasePath(soldOut.ID), "not-a-jwt"},
				{"event tidak ada", purchasePath(missingEventID), token},
				{"stok habis", purchasePath(soldOut.ID), token},
				{"masa jual belum dibuka", purchasePath(notOpen.ID), token},
				{"id bukan angka", purchasePath("abc"), token},
				{"id di luar jangkauan int4", purchasePath("99999999999999"), token},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					_, env := call(t, http.MethodPost, tc.path, tc.token, nil)
					assertNoInternalLeak(t, env)
				})
			}
		})
	})
}
