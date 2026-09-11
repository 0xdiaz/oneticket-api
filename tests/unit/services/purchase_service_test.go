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

// seedPurchasableEvent inserts an event whose sale is already open and returns
// it. priceCents is a parameter because the money path is what several of the
// cases below are about.
func seedPurchasableEvent(repo *mocks.MockEventRepository, priceCents int64) *models.Event {
	now := time.Now()
	return repo.Seed(&models.Event{
		Name:         "Flash Sale Demo",
		Venue:        "Demo Hall",
		StartsAt:     now.Add(24 * time.Hour),
		SaleStartsAt: now.Add(-time.Minute),
		PriceCents:   priceCents,
		TotalTickets: 100,
	})
}

// TestPurchaseService_Purchase pins the service's behaviour against fakes: no
// database, no transaction, no HTTP. What is left is exactly the domain rules
// the service owns, which sentinel each one produces, and what it hands the
// repository.
func TestPurchaseService_Purchase(t *testing.T) {
	t.Run("positive", purchaseUnitPositive)
	t.Run("negative", purchaseUnitNegative)
	t.Run("edge case", purchaseUnitEdgeCases)
}

func purchaseUnitPositive(t *testing.T) {
	t.Run("mengembalikan tiket yang diklaim beserta id purchase", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		event := seedPurchasableEvent(eventRepo, 15000000)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotZero(t, got.ID, "purchase id must come back to the caller")
		assert.NotZero(t, got.TicketID, "the claimed seat must be identified")
		assert.NotEmpty(t, got.TicketCode, "the seat code is what the buyer shows at the door")
		assert.Equal(t, event.ID, got.EventID)
	})

	t.Run("menyalin harga event ke dalam purchase", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		event := seedPurchasableEvent(eventRepo, 15000000)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(15000000), got.PriceCents)

		// The snapshot has to be the event's price at purchase time, so
		// assert on what the repository was actually handed, not only on
		// what came back.
		require.Len(t, purchaseRepo.Calls, 1)
		assert.Equal(t, int64(15000000), purchaseRepo.Calls[0].PriceCents)
		assert.Equal(t, event.ID, purchaseRepo.Calls[0].EventID)
		assert.Equal(t, uint(7), purchaseRepo.Calls[0].UserID)
	})

	t.Run("dua pembelian user yang sama menghasilkan tiket berbeda", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		event := seedPurchasableEvent(eventRepo, 15000000)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)

		first, err := service.Purchase(context.Background(), event.ID, 7)
		require.NoError(t, err, "the first purchase must not be rejected")
		require.NotNil(t, first)

		second, err := service.Purchase(context.Background(), event.ID, 7)
		require.NoError(t, err, "one user buying twice is allowed; there is no per-user limit")
		require.NotNil(t, second)

		assert.NotEqual(t, first.TicketID, second.TicketID, "two purchases must be two seats")
	})
}

func purchaseUnitNegative(t *testing.T) {
	t.Run("event tidak dikenal mengembalikan ErrEventNotFound", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), 999, 7)

		assert.Nil(t, got)
		assert.True(t, errors.Is(err, services.ErrEventNotFound), "expected ErrEventNotFound, got %v", err)
	})

	t.Run("stok habis mengembalikan ErrNoTicketsAvailable", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		purchaseRepo.NoStock = true
		event := seedPurchasableEvent(eventRepo, 15000000)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		assert.Nil(t, got)
		assert.True(t, errors.Is(err, services.ErrNoTicketsAvailable), "expected ErrNoTicketsAvailable, got %v", err)
	})

	t.Run("masa jual belum dibuka mengembalikan ErrSaleNotOpen", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()

		now := time.Now()
		event := eventRepo.Seed(&models.Event{
			Name:         "Not Open Yet",
			Venue:        "Demo Hall",
			StartsAt:     now.Add(48 * time.Hour),
			SaleStartsAt: now.Add(time.Hour),
			PriceCents:   15000000,
			TotalTickets: 100,
		})

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		assert.Nil(t, got)
		assert.True(t, errors.Is(err, services.ErrSaleNotOpen), "expected ErrSaleNotOpen, got %v", err)
		// The check belongs before the claim: a ticket locked for a sale
		// that has not opened is inventory taken out of circulation for
		// nothing.
		assert.Equal(t, 0, purchaseRepo.CallCount(), "the claim must never be attempted")
	})

	t.Run("kegagalan repository event dibungkus bukan ditelan", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		eventRepo.GetErr = errors.New("database is down")
		purchaseRepo := mocks.NewMockPurchaseRepository()

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), 1, 7)

		assert.Nil(t, got)
		require.Error(t, err)
		// An outage is not an absent event. Reporting it as 404 hides a
		// failing database behind a client mistake.
		assert.False(t, errors.Is(err, services.ErrEventNotFound), "an outage must not be reported as a missing event")
	})

	t.Run("kegagalan repository purchase dibungkus bukan ditelan", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		purchaseRepo.ClaimErr = errors.New("deadlock detected")
		event := seedPurchasableEvent(eventRepo, 15000000)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		assert.Nil(t, got)
		require.Error(t, err)
		// A failed claim is not a sold-out event: one is a 500, the other
		// a 409, and collapsing them loses the outage.
		assert.False(t, errors.Is(err, services.ErrNoTicketsAvailable), "a claim failure must not be reported as sold out")
	})
}

func purchaseUnitEdgeCases(t *testing.T) {
	t.Run("price_cents 9007199254740993 melewati service tanpa berubah", func(t *testing.T) {
		// Above float64's exact-integer limit: any float conversion
		// anywhere on the path turns this red.
		const huge int64 = 9007199254740993

		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		event := seedPurchasableEvent(eventRepo, huge)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, huge, got.PriceCents)

		require.Len(t, purchaseRepo.Calls, 1)
		assert.Equal(t, huge, purchaseRepo.Calls[0].PriceCents)
	})

	t.Run("price_cents nol diterima", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()
		event := seedPurchasableEvent(eventRepo, 0)

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		// Zero is the lower bound of CHECK (price_cents >= 0), not an
		// error condition: a free ticket is still a ticket.
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(0), got.PriceCents)
	})

	t.Run("masa jual dibuka satu detik lalu diterima", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()

		now := time.Now()
		event := eventRepo.Seed(&models.Event{
			Name:         "Just Opened",
			Venue:        "Demo Hall",
			StartsAt:     now.Add(24 * time.Hour),
			SaleStartsAt: now.Add(-time.Second),
			PriceCents:   15000000,
			TotalTickets: 100,
		})

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.False(t, errors.Is(err, services.ErrSaleNotOpen))
	})

	t.Run("masa jual dibuka satu detik lagi ditolak", func(t *testing.T) {
		eventRepo := mocks.NewMockEventRepository()
		purchaseRepo := mocks.NewMockPurchaseRepository()

		now := time.Now()
		event := eventRepo.Seed(&models.Event{
			Name:         "Opens In A Moment",
			Venue:        "Demo Hall",
			StartsAt:     now.Add(24 * time.Hour),
			SaleStartsAt: now.Add(time.Second),
			PriceCents:   15000000,
			TotalTickets: 100,
		})

		service := services.NewPurchaseService(eventRepo, purchaseRepo)
		got, err := service.Purchase(context.Background(), event.ID, 7)

		assert.Nil(t, got)
		assert.True(t, errors.Is(err, services.ErrSaleNotOpen), "expected ErrSaleNotOpen, got %v", err)
	})
}
