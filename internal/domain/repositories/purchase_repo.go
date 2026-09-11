package repositories

import (
	"errors"

	"github.com/0xdiaz/oneticket-api/internal/domain/models"
)

// ErrNoAvailableTicket reports that an event had no ticket left to claim.
//
// It is a condition, not a failure: the claim statement simply affected no
// rows. The repository says what happened; the service decides what it means
// to the caller, the same division the other repositories here follow when
// they translate a missing row into (nil, nil).
var ErrNoAvailableTicket = errors.New("no available ticket")

// ClaimResult is what one successful claim produced: the seat that was taken
// and the ownership row written for it.
type ClaimResult struct {
	Purchase *models.Purchase
	Ticket   *models.Ticket
}

// PurchaseRepository is the only writer of tickets.status and purchases.
//
// Both writes belong to one transaction. Splitting them across two calls would
// put the transaction boundary in the service, which cannot hold one: the
// database handle is package-level here, not injected.
type PurchaseRepository interface {
	// ClaimTicket marks the lowest-id available ticket of an event sold and
	// records ownership of it, atomically.
	//
	// priceCents is passed in rather than read here so the snapshot comes from
	// the event the service already loaded and validated.
	//
	// Returns ErrNoAvailableTicket when the event has no available ticket.
	ClaimTicket(eventID, userID uint, priceCents int64) (*ClaimResult, error)
}

type purchaseRepo struct{}

// NewPurchaseRepository returns a new PurchaseRepository implementation.
func NewPurchaseRepository() PurchaseRepository {
	return &purchaseRepo{}
}

func (r *purchaseRepo) ClaimTicket(eventID, userID uint, priceCents int64) (*ClaimResult, error) {
	panic("not implemented: U7")
}
