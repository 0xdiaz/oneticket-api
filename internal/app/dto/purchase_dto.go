package dto

import "time"

// PurchaseResponse is the API representation of a completed purchase.
//
// There is no user_id field on purpose: the owner is whoever made the request,
// so returning it adds nothing a client does not already know.
type PurchaseResponse struct {
	ID      uint `json:"id"`
	EventID uint `json:"event_id"`

	TicketID uint `json:"ticket_id"`

	// TicketCode is the seat identifier, unique within one event.
	TicketCode string `json:"ticket_code"`

	// PriceCents is what was paid, in the smallest currency unit. It stays an
	// integer all the way to the client; formatting is the client's job.
	PriceCents int64 `json:"price_cents"`

	PurchasedAt time.Time `json:"purchased_at"`
}
