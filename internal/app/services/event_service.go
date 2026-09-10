package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/app/dto"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
)

// ErrEventNotFound is returned when the requested event does not exist.
var ErrEventNotFound = errors.New("event not found")

// EventService handles event and ticket availability business logic.
type EventService struct {
	eventRepo  repositories.EventRepository
	ticketRepo repositories.TicketRepository
}

// NewEventService creates a new EventService instance.
func NewEventService(eventRepo repositories.EventRepository, ticketRepo repositories.TicketRepository) *EventService {
	return &EventService{eventRepo: eventRepo, ticketRepo: ticketRepo}
}

// List returns every event with its live ticket availability.
//
// Availability is counted per event, so this issues one query per event on
// top of the list query. Fine at demo scale, and deliberately left that way:
// it is the honest starting point for a later performance pass.
func (s *EventService) List(ctx context.Context) (responses []*dto.EventResponse, err error) {
	ctx, start := logger.LogStart(ctx, "EventService.List")

	events, err := s.eventRepo.List()
	if err != nil {
		logger.LogFinish(ctx, "EventService.List", err, start)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	responses = make([]*dto.EventResponse, 0, len(events))
	for _, event := range events {
		response, buildErr := s.buildResponse(event)
		if buildErr != nil {
			logger.LogFinish(ctx, "EventService.List", buildErr, start)
			return nil, buildErr
		}
		responses = append(responses, response)
	}

	logger.LogFinish(ctx, "EventService.List", nil, start)
	return responses, nil
}

// Get returns a single event with its live ticket availability.
//
// Returns ErrEventNotFound when no event has the given id.
func (s *EventService) Get(ctx context.Context, id uint) (response *dto.EventResponse, err error) {
	ctx, start := logger.LogStart(ctx, "EventService.Get")

	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		logger.LogFinish(ctx, "EventService.Get", err, start)
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		logger.LogFinish(ctx, "EventService.Get", ErrEventNotFound, start)
		return nil, ErrEventNotFound
	}

	response, err = s.buildResponse(event)
	if err != nil {
		logger.LogFinish(ctx, "EventService.Get", err, start)
		return nil, err
	}

	logger.LogFinish(ctx, "EventService.Get", nil, start)
	return response, nil
}

// buildResponse assembles the API representation of one event, counting its
// currently available tickets.
func (s *EventService) buildResponse(event *models.Event) (*dto.EventResponse, error) {
	available, err := s.ticketRepo.CountByStatus(event.ID, models.TicketStatusAvailable)
	if err != nil {
		return nil, fmt.Errorf("failed to count available tickets: %w", err)
	}

	return &dto.EventResponse{
		ID:               event.ID,
		Name:             event.Name,
		Venue:            event.Venue,
		StartsAt:         event.StartsAt,
		SaleStartsAt:     event.SaleStartsAt,
		PriceCents:       event.PriceCents,
		TotalTickets:     event.TotalTickets,
		AvailableTickets: available,
		SaleOpen:         !event.SaleStartsAt.After(time.Now()),
	}, nil
}
