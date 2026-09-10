package repositories

import (
	"errors"
	"fmt"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNoTicketAvailable is returned when an event has no ticket left to sell.
// It is a business outcome, not a failure: the caller turns it into a 409.
var ErrNoTicketAvailable = errors.New("no ticket available")

// OrderRepository defines data access for purchases.
//
// PurchaseTicket is deliberately a single method rather than a set of smaller
// ones: claiming a ticket and recording its order must happen together, so the
// repository owns the whole write. A caller cannot leave a ticket sold with no
// order behind it.
type OrderRepository interface {
	// PurchaseTicket claims one available ticket for the event and records the
	// order against it, atomically.
	//
	// priceCents is passed in rather than read here: the caller froze it when
	// the purchase started, and the order must store that value.
	//
	// Returns ErrNoTicketAvailable when the event is sold out.
	PurchaseTicket(eventID, userID uint, priceCents int64) (*models.Order, *models.Ticket, error)
}

type orderRepo struct{}

// NewOrderRepository returns a new OrderRepository implementation.
func NewOrderRepository() OrderRepository {
	return &orderRepo{}
}

func (r *orderRepo) PurchaseTicket(eventID, userID uint, priceCents int64) (*models.Order, *models.Ticket, error) {
	var ticket models.Ticket
	var order models.Order

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Claim a ticket for this transaction alone. SKIP LOCKED steps over
		// rows another buyer is already holding instead of queueing behind
		// them, so concurrent buyers take different seats rather than
		// serialising on the same one.
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("event_id = ? AND status = ?", eventID, models.TicketStatusAvailable).
			Order("code ASC").
			First(&ticket).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoTicketAvailable
			}
			logger.Errorf("failed to claim ticket: %v", err)
			return fmt.Errorf("failed to claim ticket: %w", err)
		}

		if err := tx.Model(&models.Ticket{}).
			Where("id = ?", ticket.ID).
			Update("status", models.TicketStatusSold).Error; err != nil {
			logger.Errorf("failed to mark ticket sold: %v", err)
			return fmt.Errorf("failed to mark ticket sold: %w", err)
		}
		ticket.Status = models.TicketStatusSold

		order = models.Order{
			UserID:     userID,
			EventID:    eventID,
			TicketID:   ticket.ID,
			PriceCents: priceCents,
			Status:     models.OrderStatusPaid,
		}
		if err := tx.Create(&order).Error; err != nil {
			logger.Errorf("failed to record order: %v", err)
			return fmt.Errorf("failed to record order: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return &order, &ticket, nil
}
