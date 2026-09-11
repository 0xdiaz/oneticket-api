package models

import "time"

// Purchase is the ownership record for one sold Ticket.
//
// Ownership lives here rather than as a column on Ticket: tickets.status
// tracks stock, purchases tracks who holds the seat. The cost of that split is
// that a sale is two writes, so PurchaseRepository performs both inside one
// transaction. A ticket marked sold without a matching purchases row is
// inventory lost for good, and no test that only reads status would see it.
type Purchase struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// TicketID is unique across the table: one ticket can only be bought once.
	// Kept in sync by hand with UNIQUE in
	// migrations/sql/000007_create_purchases_table.up.sql.
	TicketID uint `json:"ticket_id" gorm:"not null;uniqueIndex"`

	UserID uint `json:"user_id" gorm:"not null"`

	// PriceCents is the event price at the moment of sale, in the smallest
	// currency unit. A snapshot, not a lookup: repricing the event later does
	// not rewrite what someone paid. Never a float, at any layer.
	PriceCents int64 `json:"price_cents" gorm:"not null"`

	// PurchasedAt is filled by the database default rather than by the
	// application, so the timestamp comes from one clock.
	PurchasedAt time.Time `json:"purchased_at" gorm:"not null;default:now()"`
}

// TableName specifies the database table name for Purchase model.
func (p *Purchase) TableName() string {
	return "purchases"
}
