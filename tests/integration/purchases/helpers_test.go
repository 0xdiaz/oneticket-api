// Fixtures and read-back helpers shared by the purchase integration tests.
//
// Kept out of purchase_flow_test.go so neither file crosses the 300-line
// limit in docs/00_AI_CRITICAL_RULES.md, and so the concurrency test can use
// the same seeds without duplicating them.
package purchases_test

import (
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
	"github.com/stretchr/testify/require"
)

// insertEvent writes one event row with the given sale start and no tickets.
func insertEvent(t *testing.T, name string, total int, saleStartsAt time.Time) *models.Event {
	t.Helper()

	event := &models.Event{
		Name:         name,
		Venue:        "Integration Hall",
		StartsAt:     time.Now().Add(24 * time.Hour),
		SaleStartsAt: saleStartsAt,
		PriceCents:   15000000,
		TotalTickets: total,
	}
	require.NoError(t, database.DB.Create(event).Error)
	return event
}

// insertTickets generates n available tickets for an event.
func insertTickets(t *testing.T, eventID uint, n int) {
	t.Helper()
	if n == 0 {
		return
	}

	rows := make([]models.Ticket, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, models.Ticket{
			EventID: eventID,
			Code:    codeFor(i),
			Status:  models.TicketStatusAvailable,
		})
	}
	require.NoError(t, database.DB.Create(&rows).Error)
}

func codeFor(n int) string {
	const digits = "0123456789"
	return "A-" + string([]byte{digits[(n/100)%10], digits[(n/10)%10], digits[n%10]})
}

// seedEvent inserts an event whose sale is already open with n available
// tickets.
func seedEvent(t *testing.T, name string, tickets int) *models.Event {
	t.Helper()
	event := insertEvent(t, name, tickets, time.Now().Add(-time.Minute))
	insertTickets(t, event.ID, tickets)
	return event
}

// seedUser inserts a real users row. purchases.user_id carries a foreign key
// and harness.Reset truncates users, so every test needs one of these.
func seedUser(t *testing.T) *models.User {
	t.Helper()

	user := &models.User{
		Name:     "Integration Buyer",
		Email:    "buyer@example.local",
		Password: "$2a$10$integrationtesthashnotarealpasswordvaluexxxxxxxxxxxxx",
	}
	require.NoError(t, database.DB.Create(user).Error)
	return user
}

// seedSale resets the database and seeds an open sale plus its buyer.
func seedSale(t *testing.T, name string, tickets int) (*models.Event, *models.User) {
	t.Helper()
	harness.Reset(t)
	return seedEvent(t, name, tickets), seedUser(t)
}

func newService() *services.PurchaseService {
	return services.NewPurchaseService(
		repositories.NewEventRepository(),
		repositories.NewPurchaseRepository(),
	)
}

func newEventService() *services.EventService {
	return services.NewEventService(
		repositories.NewEventRepository(),
		repositories.NewTicketRepository(),
	)
}

// lowestAvailable returns the available ticket of an event with the smallest id.
func lowestAvailable(t *testing.T, eventID uint) *models.Ticket {
	t.Helper()

	var ticket models.Ticket
	require.NoError(t, database.DB.
		Where("event_id = ? AND status = ?", eventID, models.TicketStatusAvailable).
		Order("id ASC").First(&ticket).Error)
	return &ticket
}

// purchaseRow reads the purchases row written for a ticket.
func purchaseRow(t *testing.T, ticketID uint) *models.Purchase {
	t.Helper()

	var row models.Purchase
	require.NoError(t, database.DB.Where("ticket_id = ?", ticketID).First(&row).Error)
	return &row
}

// countPurchases counts every row in purchases.
func countPurchases(t *testing.T) int64 {
	t.Helper()

	var n int64
	require.NoError(t, database.DB.Model(&models.Purchase{}).Count(&n).Error)
	return n
}

// ticketStatus reads one ticket's status straight from the database.
func ticketStatus(t *testing.T, ticketID uint) string {
	t.Helper()

	var ticket models.Ticket
	require.NoError(t, database.DB.First(&ticket, ticketID).Error)
	return ticket.Status
}

// TestPurchaseService_Purchase drives the whole checkout path against a real
// Postgres: the service, the repository transaction, and the constraints the
// migration installed. Nothing here is mocked, so a disagreement between any
// two of those three shows up as a failure rather than as a green mock.
