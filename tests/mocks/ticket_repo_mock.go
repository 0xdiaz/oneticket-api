package mocks

import (
	"sort"
	"sync"

	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
)

// MockTicketRepository is an in-memory TicketRepository for unit tests.
type MockTicketRepository struct {
	mu      sync.RWMutex
	nextID  uint
	tickets []*models.Ticket

	// CountErr and ListErr, when set, are returned instead of data.
	CountErr error
	ListErr  error
}

// NewMockTicketRepository returns a new MockTicketRepository.
func NewMockTicketRepository() *MockTicketRepository {
	return &MockTicketRepository{}
}

// Ensure MockTicketRepository implements repositories.TicketRepository.
var _ repositories.TicketRepository = (*MockTicketRepository)(nil)

// Seed inserts a ticket directly, assigning an ID when it has none.
func (m *MockTicketRepository) Seed(ticket *models.Ticket) *models.Ticket {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ticket.ID == 0 {
		m.nextID++
		ticket.ID = m.nextID
	}
	m.tickets = append(m.tickets, ticket)
	return ticket
}

// SeedAvailable inserts n available tickets for an event, coded A-001 upward.
func (m *MockTicketRepository) SeedAvailable(eventID uint, n int) {
	for i := 1; i <= n; i++ {
		m.Seed(&models.Ticket{
			EventID: eventID,
			Code:    ticketCode(i),
			Status:  models.TicketStatusAvailable,
		})
	}
}

func (m *MockTicketRepository) CountByStatus(eventID uint, status string) (int64, error) {
	if m.CountErr != nil {
		return 0, m.CountErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, ticket := range m.tickets {
		if ticket.EventID == eventID && ticket.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *MockTicketRepository) ListByEvent(eventID uint) ([]*models.Ticket, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	tickets := make([]*models.Ticket, 0)
	for _, ticket := range m.tickets {
		if ticket.EventID == eventID {
			tickets = append(tickets, ticket)
		}
	}
	sort.Slice(tickets, func(i, j int) bool { return tickets[i].Code < tickets[j].Code })
	return tickets, nil
}

// ticketCode renders the seat code for the nth ticket, matching the seeder.
func ticketCode(n int) string {
	const digits = "0123456789"
	return "A-" + string([]byte{
		digits[(n/100)%10],
		digits[(n/10)%10],
		digits[n%10],
	})
}
