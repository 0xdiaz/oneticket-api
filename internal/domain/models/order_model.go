package models

import "time"

// Order status values. Kept in sync with the CHECK constraint in
// migrations/sql/000007_create_orders_table.up.sql.
const (
	OrderStatusPaid     = "paid"
	OrderStatusRefunded = "refunded"
)

// Order is the record of one completed ticket purchase.
//
// It exists separately from Ticket because a purchase has facts a seat does
// not: who bought it and what they paid at the time. PriceCents is frozen
// here, so a later change to the event price never rewrites history.
type Order struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	UserID   uint `json:"user_id" gorm:"not null;index"`
	EventID  uint `json:"event_id" gorm:"not null"`
	TicketID uint `json:"ticket_id" gorm:"not null;uniqueIndex"`

	// PriceCents is what the buyer paid, in the smallest currency unit.
	PriceCents int64 `json:"price_cents" gorm:"not null"`

	// Status is one of OrderStatusPaid or OrderStatusRefunded.
	Status string `json:"status" gorm:"type:varchar(16);not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the database table name for Order model.
func (o *Order) TableName() string {
	return "orders"
}
