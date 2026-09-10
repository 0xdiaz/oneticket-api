package orders_test

import (
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/stretchr/testify/require"
)

// seedEvent inserts an event with n available tickets.
func seedEvent(t *testing.T, name string, tickets int, priceCents int64) *models.Event {
	t.Helper()

	now := time.Now()
	event := &models.Event{
		Name: name, Venue: "Integration Hall",
		StartsAt: now.Add(24 * time.Hour), SaleStartsAt: now.Add(-time.Minute),
		PriceCents: priceCents, TotalTickets: tickets,
	}
	require.NoError(t, database.DB.Create(event).Error)

	rows := make([]models.Ticket, 0, tickets)
	for i := 1; i <= tickets; i++ {
		rows = append(rows, models.Ticket{
			EventID: event.ID, Code: ticketCode(i), Status: models.TicketStatusAvailable,
		})
	}
	if tickets > 0 {
		require.NoError(t, database.DB.Create(&rows).Error)
	}
	return event
}

// seedUser inserts a buyer.
func seedUser(t *testing.T, email string) *models.User {
	t.Helper()
	user := &models.User{Name: "Buyer", Email: email, Password: "x"}
	require.NoError(t, database.DB.Create(user).Error)
	return user
}

func ticketCode(n int) string {
	const digits = "0123456789"
	return "A-" + string([]byte{digits[(n/100)%10], digits[(n/10)%10], digits[n%10]})
}

// countTickets returns how many tickets of an event are in the given status.
func countTickets(t *testing.T, eventID uint, status string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, database.DB.Model(&models.Ticket{}).
		Where("event_id = ? AND status = ?", eventID, status).Count(&n).Error)
	return n
}
