package mocks

import (
	"sort"
	"sync"

	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
)

// MockEventRepository is an in-memory EventRepository for unit tests.
type MockEventRepository struct {
	mu     sync.RWMutex
	nextID uint
	byID   map[uint]*models.Event

	// ListErr and GetErr, when set, are returned instead of data.
	ListErr error
	GetErr  error
}

// NewMockEventRepository returns a new MockEventRepository.
func NewMockEventRepository() *MockEventRepository {
	return &MockEventRepository{byID: make(map[uint]*models.Event)}
}

// Ensure MockEventRepository implements repositories.EventRepository.
var _ repositories.EventRepository = (*MockEventRepository)(nil)

// Seed inserts an event directly, assigning an ID when it has none.
func (m *MockEventRepository) Seed(event *models.Event) *models.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byID == nil {
		m.byID = make(map[uint]*models.Event)
	}
	if event.ID == 0 {
		m.nextID++
		event.ID = m.nextID
	}
	m.byID[event.ID] = event
	return event
}

func (m *MockEventRepository) List() ([]*models.Event, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]*models.Event, 0, len(m.byID))
	for _, event := range m.byID {
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].SaleStartsAt.Before(events[j].SaleStartsAt)
	})
	return events, nil
}

func (m *MockEventRepository) GetByID(id uint) (*models.Event, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	event, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return event, nil
}
