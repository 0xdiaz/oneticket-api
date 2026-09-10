package repositories

import (
	"errors"
	"fmt"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
	"gorm.io/gorm"
)

// EventRepository defines data access for events.
type EventRepository interface {
	// List returns every event, soonest sale first.
	List() ([]*models.Event, error)
	// GetByID returns the event with the given id, or (nil, nil) when it does
	// not exist. A missing row is not an error at this layer.
	GetByID(id uint) (*models.Event, error)
}

type eventRepo struct{}

// NewEventRepository returns a new EventRepository implementation.
func NewEventRepository() EventRepository {
	return &eventRepo{}
}

func (r *eventRepo) List() ([]*models.Event, error) {
	var events []*models.Event
	if err := database.DB.Order("sale_starts_at ASC").Find(&events).Error; err != nil {
		logger.Errorf("failed to list events: %v", err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	return events, nil
}

func (r *eventRepo) GetByID(id uint) (*models.Event, error) {
	var event models.Event
	err := database.DB.Where("id = ?", id).First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("failed to get event by id: %v", err)
		return nil, fmt.Errorf("failed to get event by id: %w", err)
	}
	return &event, nil
}
