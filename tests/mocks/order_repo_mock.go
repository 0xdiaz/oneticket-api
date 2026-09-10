package mocks

import (
	"sync"

	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
)

// MockOrderRepository is an in-memory OrderRepository for unit tests.
type MockOrderRepository struct {
	mu     sync.Mutex
	nextID uint

	// Orders records every purchase this fake accepted, in order.
	Orders []*models.Order

	// SoldOut makes PurchaseTicket report the event as sold out.
	SoldOut bool
	// PurchaseErr, when set, is returned instead of a purchase. It takes
	// precedence over SoldOut so an infrastructure failure can be simulated
	// separately from the business outcome.
	PurchaseErr error
}

// NewMockOrderRepository returns a new MockOrderRepository.
func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{}
}

// Compile-time proof the fake still matches the real interface.
var _ repositories.OrderRepository = (*MockOrderRepository)(nil)

func (m *MockOrderRepository) PurchaseTicket(eventID, userID uint, priceCents int64) (*models.Order, *models.Ticket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.PurchaseErr != nil {
		return nil, nil, m.PurchaseErr
	}
	if m.SoldOut {
		return nil, nil, repositories.ErrNoTicketAvailable
	}

	m.nextID++
	ticket := &models.Ticket{
		ID:      m.nextID,
		EventID: eventID,
		Code:    ticketCode(int(m.nextID)),
		Status:  models.TicketStatusSold,
	}
	order := &models.Order{
		ID:         m.nextID,
		UserID:     userID,
		EventID:    eventID,
		TicketID:   ticket.ID,
		PriceCents: priceCents,
		Status:     models.OrderStatusPaid,
	}
	m.Orders = append(m.Orders, order)
	return order, ticket, nil
}
