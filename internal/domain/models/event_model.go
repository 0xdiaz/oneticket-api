package models

import "time"

// Event is a sellable event.
//
// Tickets are pre-generated as individual rows in the tickets table rather
// than tracked as a counter here, so every sold ticket keeps its own identity
// (its code) and can be traced back to exactly one buyer.
type Event struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"type:varchar(200);not null"`
	Venue string `json:"venue" gorm:"type:varchar(200);not null"`

	StartsAt     time.Time `json:"starts_at" gorm:"not null"`
	SaleStartsAt time.Time `json:"sale_starts_at" gorm:"not null;index"`

	// PriceCents is the ticket price in the smallest currency unit.
	// Money is an integer here on purpose: floats lose cents under arithmetic.
	PriceCents int64 `json:"price_cents" gorm:"not null"`

	// TotalTickets is the planned inventory, i.e. how many ticket rows were
	// generated for this event. It is NOT live availability: availability is
	// derived by counting tickets whose status is TicketStatusAvailable.
	TotalTickets int `json:"total_tickets" gorm:"not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the database table name for Event model.
func (e *Event) TableName() string {
	return "events"
}
