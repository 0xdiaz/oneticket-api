package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xdiaz/oneticket-api/internal/app/dto"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
)

// ErrSoldOut is returned when an event has no ticket left to sell.
var ErrSoldOut = errors.New("event is sold out")

// OrderService handles ticket purchase business logic.
type OrderService struct {
	eventRepo repositories.EventRepository
	orderRepo repositories.OrderRepository
}

// NewOrderService creates a new OrderService instance.
func NewOrderService(eventRepo repositories.EventRepository, orderRepo repositories.OrderRepository) *OrderService {
	return &OrderService{eventRepo: eventRepo, orderRepo: orderRepo}
}

// PurchaseTicket sells one ticket of the event to the given user.
//
// The event price is read here and handed to the repository, which is what
// freezes it onto the order: a later change to the event never rewrites a
// purchase that already happened.
//
// Returns ErrEventNotFound when the event does not exist, and ErrSoldOut when
// it has no available ticket left.
func (s *OrderService) PurchaseTicket(ctx context.Context, eventID, userID uint) (response *dto.PurchaseResponse, err error) {
	ctx, start := logger.LogStart(ctx, "OrderService.PurchaseTicket")

	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		logger.LogFinish(ctx, "OrderService.PurchaseTicket", err, start)
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		logger.LogFinish(ctx, "OrderService.PurchaseTicket", ErrEventNotFound, start)
		return nil, ErrEventNotFound
	}

	order, ticket, err := s.orderRepo.PurchaseTicket(eventID, userID, event.PriceCents)
	if err != nil {
		// Only an empty inventory becomes ErrSoldOut. An infrastructure
		// failure must stay a failure: reporting it as "sold out" would tell
		// the buyer a comfortable lie and hide an outage.
		if errors.Is(err, repositories.ErrNoTicketAvailable) {
			logger.LogFinish(ctx, "OrderService.PurchaseTicket", ErrSoldOut, start)
			return nil, ErrSoldOut
		}
		logger.LogFinish(ctx, "OrderService.PurchaseTicket", err, start)
		return nil, fmt.Errorf("failed to purchase ticket: %w", err)
	}

	logger.Infof("ticket sold: event=%d ticket=%s user=%d", eventID, ticket.Code, userID)
	logger.LogFinish(ctx, "OrderService.PurchaseTicket", nil, start)

	return &dto.PurchaseResponse{
		OrderID:    order.ID,
		TicketID:   ticket.ID,
		TicketCode: ticket.Code,
		EventID:    eventID,
		PriceCents: order.PriceCents,
	}, nil
}
