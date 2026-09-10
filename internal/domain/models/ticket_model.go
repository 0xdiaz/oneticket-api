package models

import "time"

// Ticket status values. Kept in sync with the CHECK constraint in
// migrations/sql/000006_create_tickets_table.up.sql.
const (
	TicketStatusAvailable = "available"
	TicketStatusSold      = "sold"
)

// Ticket is one individual, sellable seat for an Event.
//
// Buyer linkage — which order or user a sold ticket belongs to — is
// deliberately not modelled yet; it arrives together with the checkout flow.
type Ticket struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	EventID uint `json:"event_id" gorm:"not null;index"`

	// Code is the human-readable ticket identifier, unique within one event.
	Code string `json:"code" gorm:"type:varchar(32);not null"`

	// Status is one of TicketStatusAvailable or TicketStatusSold.
	Status string `json:"status" gorm:"type:varchar(16);not null;index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the database table name for Ticket model.
func (t *Ticket) TableName() string {
	return "tickets"
}
