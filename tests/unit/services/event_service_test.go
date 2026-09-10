package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventServiceList(t *testing.T) {
	eventRepo := mocks.NewMockEventRepository()
	ticketRepo := mocks.NewMockTicketRepository()

	now := time.Now()
	later := eventRepo.Seed(&models.Event{
		Name: "Later Sale", Venue: "Hall B",
		StartsAt: now.Add(48 * time.Hour), SaleStartsAt: now.Add(2 * time.Hour),
		PriceCents: 20000000, TotalTickets: 50,
	})
	sooner := eventRepo.Seed(&models.Event{
		Name: "Sooner Sale", Venue: "Hall A",
		StartsAt: now.Add(24 * time.Hour), SaleStartsAt: now.Add(-1 * time.Hour),
		PriceCents: 15000000, TotalTickets: 100,
	})
	ticketRepo.SeedAvailable(sooner.ID, 3)
	ticketRepo.SeedAvailable(later.ID, 1)

	service := services.NewEventService(eventRepo, ticketRepo)
	got, err := service.List(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 2)

	// Ordered by sale start, soonest first.
	assert.Equal(t, "Sooner Sale", got[0].Name)
	assert.Equal(t, "Later Sale", got[1].Name)

	assert.Equal(t, int64(3), got[0].AvailableTickets)
	assert.Equal(t, 100, got[0].TotalTickets)
	assert.True(t, got[0].SaleOpen, "sale started an hour ago")
	assert.False(t, got[1].SaleOpen, "sale starts in two hours")
}

func TestEventServiceGet(t *testing.T) {
	eventRepo := mocks.NewMockEventRepository()
	ticketRepo := mocks.NewMockTicketRepository()

	now := time.Now()
	event := eventRepo.Seed(&models.Event{
		Name: "Flash Sale Demo", Venue: "Demo Hall",
		StartsAt: now.Add(24 * time.Hour), SaleStartsAt: now.Add(-time.Minute),
		PriceCents: 15000000, TotalTickets: 100,
	})
	ticketRepo.SeedAvailable(event.ID, 100)

	service := services.NewEventService(eventRepo, ticketRepo)

	t.Run("existing event", func(t *testing.T) {
		got, err := service.Get(context.Background(), event.ID)

		require.NoError(t, err)
		assert.Equal(t, event.ID, got.ID)
		assert.Equal(t, "Flash Sale Demo", got.Name)
		assert.Equal(t, int64(100), got.AvailableTickets)
		// Money stays an integer end to end.
		assert.Equal(t, int64(15000000), got.PriceCents)
	})

	t.Run("unknown event", func(t *testing.T) {
		got, err := service.Get(context.Background(), 999)

		assert.Nil(t, got)
		assert.True(t, errors.Is(err, services.ErrEventNotFound), "expected ErrEventNotFound, got %v", err)
	})
}

// TestEventServiceAvailabilityExcludesSold guards the meaning of
// AvailableTickets: it is counted from ticket rows, never derived from
// TotalTickets, so a sold ticket disappears from availability.
func TestEventServiceAvailabilityExcludesSold(t *testing.T) {
	eventRepo := mocks.NewMockEventRepository()
	ticketRepo := mocks.NewMockTicketRepository()

	event := eventRepo.Seed(&models.Event{
		Name: "Flash Sale Demo", Venue: "Demo Hall",
		StartsAt: time.Now().Add(24 * time.Hour), SaleStartsAt: time.Now().Add(-time.Minute),
		PriceCents: 15000000, TotalTickets: 3,
	})
	ticketRepo.SeedAvailable(event.ID, 2)
	ticketRepo.Seed(&models.Ticket{EventID: event.ID, Code: "A-003", Status: models.TicketStatusSold})

	service := services.NewEventService(eventRepo, ticketRepo)
	got, err := service.Get(context.Background(), event.ID)

	require.NoError(t, err)
	assert.Equal(t, 3, got.TotalTickets)
	assert.Equal(t, int64(2), got.AvailableTickets)
}

// TestEventServiceGetPropagatesRepoError makes sure a repository failure is
// surfaced rather than being reported as "event not found".
func TestEventServiceGetPropagatesRepoError(t *testing.T) {
	eventRepo := mocks.NewMockEventRepository()
	eventRepo.GetErr = errors.New("database is down")

	service := services.NewEventService(eventRepo, mocks.NewMockTicketRepository())
	got, err := service.Get(context.Background(), 1)

	assert.Nil(t, got)
	require.Error(t, err)
	assert.False(t, errors.Is(err, services.ErrEventNotFound))
}
