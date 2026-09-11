package repositories

import (
	"errors"
	"fmt"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
	"gorm.io/gorm"
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

// claimTicketSQL takes the lowest-id available seat of one event and marks it
// sold in a single statement.
//
// The inner SELECT is what makes the flash sale safe. FOR UPDATE locks the
// row a buyer is about to take, and SKIP LOCKED makes a concurrent buyer step
// over that row to the next one instead of queueing behind it: N simultaneous
// requests claim N distinct seats, and none of them oversells. Reading the
// availability first and updating after would leave exactly the window this
// avoids.
//
// RETURNING carries the claimed row back, so the seat is identified by the
// same statement that took it rather than by a second lookup that a
// concurrent claim could have invalidated.
//
// updated_at is set explicitly because raw SQL bypasses GORM's autoUpdateTime;
// the column is NOT NULL DEFAULT NOW() with no trigger behind it, so without
// this line a sold ticket keeps the timestamp of its creation.
const claimTicketSQL = `
UPDATE tickets SET status = ?, updated_at = NOW()
WHERE id = (
	SELECT id FROM tickets
	WHERE event_id = ? AND status = ?
	ORDER BY id
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
RETURNING id, code`

// claimedTicket is the projection RETURNING fills: only what identifies the
// seat, since every other column is already known.
type claimedTicket struct {
	ID   uint
	Code string
}

func (r *purchaseRepo) ClaimTicket(eventID, userID uint, priceCents int64) (*ClaimResult, error) {
	var result *ClaimResult

	// One transaction for both writes. A ticket flipped to sold whose
	// purchases row then fails to insert is inventory lost for good, so the
	// UPDATE has to be undone by the same failure that stops the INSERT.
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var claimed claimedTicket
		claim := tx.Raw(claimTicketSQL,
			models.TicketStatusSold,
			eventID,
			models.TicketStatusAvailable,
		).Scan(&claimed)
		if claim.Error != nil {
			logger.Errorf("failed to claim ticket: %v", claim.Error)
			return fmt.Errorf("failed to claim ticket: %w", claim.Error)
		}
		// No row matched: the event is out of stock. A condition, not a
		// failure, so it is reported as the sentinel rather than wrapped as a
		// database error.
		if claim.RowsAffected == 0 {
			return ErrNoAvailableTicket
		}

		purchase := models.Purchase{
			TicketID:   claimed.ID,
			UserID:     userID,
			PriceCents: priceCents,
		}
		// PurchasedAt is left zero on purpose: GORM omits a zero field that
		// declares a default and reads the value back, so the timestamp comes
		// from the database clock rather than from this process.
		if err := tx.Create(&purchase).Error; err != nil {
			logger.Errorf("failed to record purchase: %v", err)
			return fmt.Errorf("failed to record purchase: %w", err)
		}

		result = &ClaimResult{
			Purchase: &purchase,
			Ticket: &models.Ticket{
				ID:      claimed.ID,
				EventID: eventID,
				Code:    claimed.Code,
				Status:  models.TicketStatusSold,
			},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
