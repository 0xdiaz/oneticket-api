package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/app/dto"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
)

// ErrNoTicketsAvailable is returned when an event has sold out.
var ErrNoTicketsAvailable = errors.New("no tickets available")

// ErrSaleNotOpen is returned when an event's sale has not started yet.
//
// Distinct from ErrNoTicketsAvailable even though both map to 409: the client
// is told which of the two it hit through the message, and the controller
// should never have to guess.
var ErrSaleNotOpen = errors.New("sale not open yet")

// PurchaseService buys one ticket for one user.
type PurchaseService struct {
	eventRepo    repositories.EventRepository
	purchaseRepo repositories.PurchaseRepository
}

// NewPurchaseService creates a new PurchaseService instance.
func NewPurchaseService(
	eventRepo repositories.EventRepository,
	purchaseRepo repositories.PurchaseRepository,
) *PurchaseService {
	return &PurchaseService{eventRepo: eventRepo, purchaseRepo: purchaseRepo}
}

// Purchase claims one available ticket of an event for a user.
//
// Returns ErrEventNotFound when the event does not exist, ErrSaleNotOpen when
// its sale has not started, and ErrNoTicketsAvailable when nothing is left.
// Choosing the HTTP status for each is the controller's job, not this one's.
func (s *PurchaseService) Purchase(ctx context.Context, eventID, userID uint) (response *dto.PurchaseResponse, err error) {
	ctx, start := logger.LogStart(ctx, "PurchaseService.Purchase")

	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		wrapped := fmt.Errorf("failed to get event: %w", err)
		logger.LogFinish(ctx, "PurchaseService.Purchase", wrapped, start)
		return nil, wrapped
	}
	if event == nil {
		logger.LogFinish(ctx, "PurchaseService.Purchase", ErrEventNotFound, start)
		return nil, ErrEventNotFound
	}

	// Checked before the claim, so a sale that has not opened never locks a
	// row. The boundary is inclusive, matching SaleOpen in EventService.
	if event.SaleStartsAt.After(time.Now()) {
		logger.LogFinish(ctx, "PurchaseService.Purchase", ErrSaleNotOpen, start)
		return nil, ErrSaleNotOpen
	}

	// The price travels with the claim so the snapshot written to purchases
	// comes from the event this call already loaded and validated, rather
	// than from a second read that could see a different price.
	claim, err := s.purchaseRepo.ClaimTicket(event.ID, userID, event.PriceCents)
	if err != nil {
		if errors.Is(err, repositories.ErrNoAvailableTicket) {
			logger.LogFinish(ctx, "PurchaseService.Purchase", ErrNoTicketsAvailable, start)
			return nil, ErrNoTicketsAvailable
		}
		wrapped := fmt.Errorf("failed to claim ticket: %w", err)
		logger.LogFinish(ctx, "PurchaseService.Purchase", wrapped, start)
		return nil, wrapped
	}

	response = &dto.PurchaseResponse{
		ID:          claim.Purchase.ID,
		EventID:     event.ID,
		TicketID:    claim.Ticket.ID,
		TicketCode:  claim.Ticket.Code,
		PriceCents:  claim.Purchase.PriceCents,
		PurchasedAt: claim.Purchase.PurchasedAt,
	}

	logger.LogFinish(ctx, "PurchaseService.Purchase", nil, start)
	return response, nil
}
