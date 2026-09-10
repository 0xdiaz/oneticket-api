package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedEvent inserts an event with n available tickets and returns it.
func seedEvent(t *testing.T, name string, tickets int) *models.Event {
	t.Helper()

	now := time.Now()
	event := &models.Event{
		Name:         name,
		Venue:        "Integration Hall",
		StartsAt:     now.Add(24 * time.Hour),
		SaleStartsAt: now.Add(-time.Minute),
		PriceCents:   15000000,
		TotalTickets: tickets,
	}
	require.NoError(t, database.DB.Create(event).Error)

	rows := make([]models.Ticket, 0, tickets)
	for i := 1; i <= tickets; i++ {
		rows = append(rows, models.Ticket{
			EventID: event.ID,
			Code:    codeFor(i),
			Status:  models.TicketStatusAvailable,
		})
	}
	if tickets > 0 {
		require.NoError(t, database.DB.Create(&rows).Error)
	}
	return event
}

func codeFor(n int) string {
	const digits = "0123456789"
	return "A-" + string([]byte{digits[(n/100)%10], digits[(n/10)%10], digits[n%10]})
}

func newService() *services.EventService {
	return services.NewEventService(
		repositories.NewEventRepository(),
		repositories.NewTicketRepository(),
	)
}

// TestEventService_Get exercises the full read path against a real database:
// migrations applied by the harness, rows written by GORM, and availability
// counted by the repository.
func TestEventService_Get(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("returns the event with live availability", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", 100)

			got, err := newService().Get(context.Background(), event.ID)

			require.NoError(t, err)
			assert.Equal(t, "Flash Sale Demo", got.Name)
			assert.Equal(t, 100, got.TotalTickets)
			assert.Equal(t, int64(100), got.AvailableTickets)
			assert.True(t, got.SaleOpen)
		})

		t.Run("availability excludes sold tickets", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Partly Sold", 10)
			require.NoError(t, database.DB.Model(&models.Ticket{}).
				Where("event_id = ? AND code = ?", event.ID, "A-001").
				Update("status", models.TicketStatusSold).Error)

			got, err := newService().Get(context.Background(), event.ID)

			require.NoError(t, err)
			assert.Equal(t, 10, got.TotalTickets)
			assert.Equal(t, int64(9), got.AvailableTickets)
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("unknown id returns ErrEventNotFound", func(t *testing.T) {
			harness.Reset(t)

			got, err := newService().Get(context.Background(), 999999)

			assert.Nil(t, got)
			assert.ErrorIs(t, err, services.ErrEventNotFound)
		})

		t.Run("database rejects an event with no inventory", func(t *testing.T) {
			harness.Reset(t)

			// The CHECK constraint in 000005_create_events_table.up.sql is the
			// thing under test here: the rule has to hold even when application
			// code forgets it.
			err := database.DB.Create(&models.Event{
				Name: "No Inventory", Venue: "Integration Hall",
				StartsAt: time.Now().Add(24 * time.Hour), SaleStartsAt: time.Now(),
				PriceCents: 1, TotalTickets: 0,
			}).Error

			require.Error(t, err)
			assert.Contains(t, err.Error(), "events_total_tickets_check")
		})

		t.Run("database rejects an unknown ticket status", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Status Check", 1)

			err := database.DB.Create(&models.Ticket{
				EventID: event.ID, Code: "X-001", Status: "refunded",
			}).Error

			require.Error(t, err)
			assert.Contains(t, err.Error(), "tickets_status_check")
		})
	})

	t.Run("edge case", func(t *testing.T) {
		t.Run("sold out event reports zero availability", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Sold Out", 3)
			require.NoError(t, database.DB.Model(&models.Ticket{}).
				Where("event_id = ?", event.ID).
				Update("status", models.TicketStatusSold).Error)

			got, err := newService().Get(context.Background(), event.ID)

			require.NoError(t, err)
			assert.Equal(t, 3, got.TotalTickets)
			assert.Equal(t, int64(0), got.AvailableTickets)
		})

		t.Run("price survives the database round trip as an integer", func(t *testing.T) {
			harness.Reset(t)
			// Larger than float64 can represent exactly, so a float column or
			// conversion anywhere on the path would corrupt it.
			const exact int64 = 9007199254740993
			event := seedEvent(t, "Precision", 1)
			require.NoError(t, database.DB.Model(&models.Event{}).
				Where("id = ?", event.ID).Update("price_cents", exact).Error)

			got, err := newService().Get(context.Background(), event.ID)

			require.NoError(t, err)
			assert.Equal(t, exact, got.PriceCents)
		})
	})
}

// TestEventService_List checks ordering and that availability is per event.
func TestEventService_List(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("orders by sale start, soonest first", func(t *testing.T) {
			harness.Reset(t)
			later := seedEvent(t, "Later", 2)
			require.NoError(t, database.DB.Model(&models.Event{}).
				Where("id = ?", later.ID).
				Update("sale_starts_at", time.Now().Add(2*time.Hour)).Error)
			seedEvent(t, "Sooner", 5)

			got, err := newService().List(context.Background())

			require.NoError(t, err)
			require.Len(t, got, 2)
			assert.Equal(t, "Sooner", got[0].Name)
			assert.Equal(t, int64(5), got[0].AvailableTickets)
			assert.Equal(t, "Later", got[1].Name)
			assert.Equal(t, int64(2), got[1].AvailableTickets)
			assert.False(t, got[1].SaleOpen)
		})
	})
}
