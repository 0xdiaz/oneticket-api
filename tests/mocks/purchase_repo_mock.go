package mocks

import (
	"sync"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
)

// ClaimCall records one ClaimTicket invocation so a test can assert on what
// the service handed the repository, not only on what came back.
type ClaimCall struct {
	EventID    uint
	UserID     uint
	PriceCents int64
}

// MockPurchaseRepository is an in-memory PurchaseRepository for unit tests.
//
// Each successful claim hands back a fresh ticket id, which is what lets a
// test show that one user buying twice gets two different seats.
type MockPurchaseRepository struct {
	mu     sync.Mutex
	nextID uint

	// NoStock makes ClaimTicket report ErrNoAvailableTicket, the way a real
	// claim reports an UPDATE that affected no rows.
	NoStock bool

	// ClaimErr, when set, is returned instead of data.
	ClaimErr error

	// Calls is the ordered record of every ClaimTicket invocation.
	Calls []ClaimCall
}

// NewMockPurchaseRepository returns a new MockPurchaseRepository.
func NewMockPurchaseRepository() *MockPurchaseRepository {
	return &MockPurchaseRepository{}
}

// Ensure MockPurchaseRepository implements repositories.PurchaseRepository.
var _ repositories.PurchaseRepository = (*MockPurchaseRepository)(nil)

func (m *MockPurchaseRepository) ClaimTicket(eventID, userID uint, priceCents int64) (*repositories.ClaimResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, ClaimCall{EventID: eventID, UserID: userID, PriceCents: priceCents})

	if m.ClaimErr != nil {
		return nil, m.ClaimErr
	}
	if m.NoStock {
		return nil, repositories.ErrNoAvailableTicket
	}

	m.nextID++
	id := m.nextID
	return &repositories.ClaimResult{
		Purchase: &models.Purchase{
			ID:          id,
			TicketID:    id,
			UserID:      userID,
			PriceCents:  priceCents,
			PurchasedAt: time.Now(),
		},
		Ticket: &models.Ticket{
			ID:      id,
			EventID: eventID,
			Code:    ticketCode(int(id)),
			Status:  models.TicketStatusSold,
		},
	}, nil
}

// CallCount returns how many times ClaimTicket was called.
func (m *MockPurchaseRepository) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Calls)
}
