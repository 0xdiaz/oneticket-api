package dto

import "time"

// EventResponse is the API representation of an event.
type EventResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Venue string `json:"venue"`

	StartsAt     time.Time `json:"starts_at"`
	SaleStartsAt time.Time `json:"sale_starts_at"`

	// PriceCents is the ticket price in the smallest currency unit. It stays
	// an integer all the way to the client; formatting is the client's job.
	PriceCents int64 `json:"price_cents"`

	// TotalTickets is the planned inventory for this event.
	TotalTickets int `json:"total_tickets"`

	// AvailableTickets is counted live from the tickets table.
	AvailableTickets int64 `json:"available_tickets"`

	// SaleOpen reports whether SaleStartsAt has passed.
	SaleOpen bool `json:"sale_open"`
}
