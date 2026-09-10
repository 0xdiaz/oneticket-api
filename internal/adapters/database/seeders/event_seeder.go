package seeders

import (
	"fmt"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
)

// demoEventTicketCount is the inventory of the seeded flash sale event.
//
// Small on purpose: a round, small number makes an oversell obvious at a
// glance — 100 tickets in, more than 100 sold out.
const demoEventTicketCount = 100

// seedDemoEvent creates one flash sale event with its ticket inventory.
//
// Idempotent: it keys off the event name, so restarting the app does not
// duplicate the event or its tickets.
func seedDemoEvent() error {
	const eventName = "Flash Sale Demo"

	var count int64
	if err := database.DB.Model(&models.Event{}).Where("name = ?", eventName).Count(&count).Error; err != nil {
		return fmt.Errorf("count events: %w", err)
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	event := models.Event{
		Name:         eventName,
		Venue:        "Demo Hall",
		StartsAt:     now.Add(30 * 24 * time.Hour),
		SaleStartsAt: now,
		// Rp150.000 expressed in the smallest currency unit.
		PriceCents:   15000000,
		TotalTickets: demoEventTicketCount,
	}
	if err := database.DB.Create(&event).Error; err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	tickets := make([]models.Ticket, 0, demoEventTicketCount)
	for i := 1; i <= demoEventTicketCount; i++ {
		tickets = append(tickets, models.Ticket{
			EventID: event.ID,
			Code:    fmt.Sprintf("A-%03d", i),
			Status:  models.TicketStatusAvailable,
		})
	}
	if err := database.DB.Create(&tickets).Error; err != nil {
		return fmt.Errorf("create tickets: %w", err)
	}

	logger.Infof("Seeded event %q (ID=%d) with %d tickets", event.Name, event.ID, len(tickets))
	return nil
}
