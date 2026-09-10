package repositories

import (
	"fmt"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
)

// TicketRepository defines data access for tickets.
//
// Only read paths live here for now. Everything the checkout needs — claiming
// a ticket, recording who bought it — is added together with that feature, so
// the locking strategy is chosen with the write path in view rather than
// guessed in advance.
type TicketRepository interface {
	// CountByStatus returns how many tickets of an event are in the given status.
	CountByStatus(eventID uint, status string) (int64, error)
	// ListByEvent returns the tickets of an event ordered by code.
	ListByEvent(eventID uint) ([]*models.Ticket, error)
}

type ticketRepo struct{}

// NewTicketRepository returns a new TicketRepository implementation.
func NewTicketRepository() TicketRepository {
	return &ticketRepo{}
}

func (r *ticketRepo) CountByStatus(eventID uint, status string) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Ticket{}).
		Where("event_id = ? AND status = ?", eventID, status).
		Count(&count).Error
	if err != nil {
		logger.Errorf("failed to count tickets by status: %v", err)
		return 0, fmt.Errorf("failed to count tickets by status: %w", err)
	}
	return count, nil
}

func (r *ticketRepo) ListByEvent(eventID uint) ([]*models.Ticket, error) {
	var tickets []*models.Ticket
	err := database.DB.Where("event_id = ?", eventID).Order("code ASC").Find(&tickets).Error
	if err != nil {
		logger.Errorf("failed to list tickets by event: %v", err)
		return nil, fmt.Errorf("failed to list tickets by event: %w", err)
	}
	return tickets, nil
}
