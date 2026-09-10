package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedPurchasableEvent registers an event the buyer can purchase from.
func seedPurchasableEvent(eventRepo *mocks.MockEventRepository, priceCents int64) *models.Event {
	return eventRepo.Seed(&models.Event{
		Name: "Flash Sale Demo", Venue: "Demo Hall",
		PriceCents: priceCents, TotalTickets: 100,
	})
}

func TestOrderServicePurchaseTicket(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("returns the purchased ticket and its order", func(t *testing.T) {
			eventRepo := mocks.NewMockEventRepository()
			orderRepo := mocks.NewMockOrderRepository()
			event := seedPurchasableEvent(eventRepo, 15000000)

			service := services.NewOrderService(eventRepo, orderRepo)
			got, err := service.PurchaseTicket(context.Background(), event.ID, 7)

			require.NoError(t, err)
			assert.NotEmpty(t, got.TicketCode)
			assert.Equal(t, event.ID, got.EventID)
			require.Len(t, orderRepo.Orders, 1, "repository called exactly once")
			assert.Equal(t, uint(7), orderRepo.Orders[0].UserID)
		})

		t.Run("freezes the event price onto the order", func(t *testing.T) {
			eventRepo := mocks.NewMockEventRepository()
			orderRepo := mocks.NewMockOrderRepository()
			event := seedPurchasableEvent(eventRepo, 15000000)

			service := services.NewOrderService(eventRepo, orderRepo)
			got, err := service.PurchaseTicket(context.Background(), event.ID, 7)

			require.NoError(t, err)
			assert.Equal(t, int64(15000000), got.PriceCents)
			assert.Equal(t, int64(15000000), orderRepo.Orders[0].PriceCents,
				"the price handed to the repository is the event price at purchase time")
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("unknown event returns ErrEventNotFound", func(t *testing.T) {
			service := services.NewOrderService(mocks.NewMockEventRepository(), mocks.NewMockOrderRepository())

			got, err := service.PurchaseTicket(context.Background(), 999, 7)

			assert.Nil(t, got)
			assert.ErrorIs(t, err, services.ErrEventNotFound)
		})

		t.Run("sold out returns ErrSoldOut", func(t *testing.T) {
			eventRepo := mocks.NewMockEventRepository()
			orderRepo := mocks.NewMockOrderRepository()
			orderRepo.SoldOut = true
			event := seedPurchasableEvent(eventRepo, 15000000)

			service := services.NewOrderService(eventRepo, orderRepo)
			got, err := service.PurchaseTicket(context.Background(), event.ID, 7)

			assert.Nil(t, got)
			assert.ErrorIs(t, err, services.ErrSoldOut)
		})

		t.Run("repository failure does not masquerade as sold out", func(t *testing.T) {
			eventRepo := mocks.NewMockEventRepository()
			orderRepo := mocks.NewMockOrderRepository()
			orderRepo.PurchaseErr = errors.New("database is down")
			event := seedPurchasableEvent(eventRepo, 15000000)

			service := services.NewOrderService(eventRepo, orderRepo)
			got, err := service.PurchaseTicket(context.Background(), event.ID, 7)

			assert.Nil(t, got)
			require.Error(t, err)
			assert.False(t, errors.Is(err, services.ErrSoldOut),
				"infrastructure failure must not be reported as an empty inventory")
			assert.False(t, errors.Is(err, repositories.ErrNoTicketAvailable))
		})

		t.Run("event lookup failure is surfaced, not swallowed", func(t *testing.T) {
			eventRepo := mocks.NewMockEventRepository()
			eventRepo.GetErr = errors.New("database is down")

			service := services.NewOrderService(eventRepo, mocks.NewMockOrderRepository())
			got, err := service.PurchaseTicket(context.Background(), 1, 7)

			assert.Nil(t, got)
			require.Error(t, err)
			assert.False(t, errors.Is(err, services.ErrEventNotFound))
		})
	})

	t.Run("edge case", func(t *testing.T) {
		t.Run("price survives the service as an exact integer", func(t *testing.T) {
			// Above float64's exact-integer limit: any float conversion on the
			// path would round this.
			const exact int64 = 9007199254740993
			eventRepo := mocks.NewMockEventRepository()
			orderRepo := mocks.NewMockOrderRepository()
			event := seedPurchasableEvent(eventRepo, exact)

			service := services.NewOrderService(eventRepo, orderRepo)
			got, err := service.PurchaseTicket(context.Background(), event.ID, 7)

			require.NoError(t, err)
			assert.Equal(t, exact, got.PriceCents)
			assert.Equal(t, exact, orderRepo.Orders[0].PriceCents)
		})

		t.Run("zero price is a valid purchase", func(t *testing.T) {
			eventRepo := mocks.NewMockEventRepository()
			orderRepo := mocks.NewMockOrderRepository()
			event := seedPurchasableEvent(eventRepo, 0)

			service := services.NewOrderService(eventRepo, orderRepo)
			got, err := service.PurchaseTicket(context.Background(), event.ID, 7)

			require.NoError(t, err)
			assert.Equal(t, int64(0), got.PriceCents)
		})
	})
}
