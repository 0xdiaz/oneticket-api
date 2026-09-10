package dto

// PurchaseResponse is the API representation of a completed purchase.
type PurchaseResponse struct {
	OrderID    uint   `json:"order_id"`
	TicketID   uint   `json:"ticket_id"`
	TicketCode string `json:"ticket_code"`
	EventID    uint   `json:"event_id"`

	// PriceCents is what the buyer paid, frozen at purchase time, in the
	// smallest currency unit.
	PriceCents int64 `json:"price_cents"`
}
