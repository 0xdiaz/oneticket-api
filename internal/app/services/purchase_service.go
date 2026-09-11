package services

import (
	"context"
	"errors"

	"github.com/0xdiaz/oneticket-api/internal/app/dto"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
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
func (s *PurchaseService) Purchase(ctx context.Context, eventID, userID uint) (*dto.PurchaseResponse, error) {
	panic("not implemented: U8")
}
