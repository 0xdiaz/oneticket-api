package purchases_test

import (
	"context"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurchaseService_Purchase(t *testing.T) {
	t.Run("positive", purchaseFlowPositive)
	t.Run("negative", purchaseFlowNegative)
	t.Run("edge case", purchaseFlowEdgeCases)
}

func purchaseFlowPositive(t *testing.T) {
	ctx := context.Background()

	t.Run("jalur penuh menulis satu baris purchases dan menandai tiketnya sold", func(t *testing.T) {
		event, user := seedSale(t, "Flash Sale Demo", 5)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		require.NoError(t, err)
		require.NotNil(t, got)
		// Verified against the database, not against the return value: the
		// response could be right while nothing was committed.
		row := purchaseRow(t, got.TicketID)
		assert.Equal(t, user.ID, row.UserID)
		assert.Equal(t, int64(1), countPurchases(t))
		assert.Equal(t, models.TicketStatusSold, ticketStatus(t, got.TicketID))
	})

	t.Run("available_tickets turun tepat satu", func(t *testing.T) {
		event, user := seedSale(t, "Countdown", 5)
		before, err := newEventService().Get(ctx, event.ID)
		require.NoError(t, err)

		_, err = newService().Purchase(ctx, event.ID, user.ID)
		require.NoError(t, err)

		after, err := newEventService().Get(ctx, event.ID)
		require.NoError(t, err)
		assert.Equal(t, before.AvailableTickets-1, after.AvailableTickets)
		assert.Equal(t, int64(4), after.AvailableTickets)
	})

	t.Run("mengklaim tiket available dengan id terkecil", func(t *testing.T) {
		event, user := seedSale(t, "Lowest Id", 5)
		// Mark the lowest id sold first. Without this a sequential scan of
		// five rows returns the smallest id even with no ORDER BY, so the
		// assertion could barely fail.
		first := lowestAvailable(t, event.ID)
		require.NoError(t, database.DB.Model(&models.Ticket{}).
			Where("id = ?", first.ID).
			Update("status", models.TicketStatusSold).Error)
		want := lowestAvailable(t, event.ID)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		require.NoError(t, err)
		assert.Equal(t, want.ID, got.TicketID)
		assert.Equal(t, want.Code, got.TicketCode)
	})

	t.Run("harga di purchases sama dengan harga event saat pembelian", func(t *testing.T) {
		event, user := seedSale(t, "Priced", 3)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		require.NoError(t, err)
		assert.Equal(t, event.PriceCents, got.PriceCents)
		assert.Equal(t, event.PriceCents, purchaseRow(t, got.TicketID).PriceCents)
	})

	t.Run("user yang sama membeli dua kali berturut-turut mendapat dua tiket berbeda", func(t *testing.T) {
		event, user := seedSale(t, "No Cap", 5)
		service := newService()

		first, err := service.Purchase(ctx, event.ID, user.ID)
		require.NoError(t, err)
		second, err := service.Purchase(ctx, event.ID, user.ID)
		require.NoError(t, err)

		assert.NotEqual(t, first.TicketID, second.TicketID)
		var mine int64
		require.NoError(t, database.DB.Model(&models.Purchase{}).
			Where("user_id = ?", user.ID).Count(&mine).Error)
		assert.Equal(t, int64(2), mine)
	})
}

func purchaseFlowNegative(t *testing.T) {
	ctx := context.Background()

	t.Run("event tidak dikenal mengembalikan ErrEventNotFound", func(t *testing.T) {
		harness.Reset(t)
		user := seedUser(t)

		got, err := newService().Purchase(ctx, 999999, user.ID)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, services.ErrEventNotFound)
	})

	t.Run("event yang seluruh tiketnya sold mengembalikan ErrNoTicketsAvailable", func(t *testing.T) {
		event, user := seedSale(t, "Sold Out", 3)
		require.NoError(t, database.DB.Model(&models.Ticket{}).
			Where("event_id = ?", event.ID).
			Update("status", models.TicketStatusSold).Error)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, services.ErrNoTicketsAvailable)
	})

	t.Run("event tanpa tiket available sejak awal mengembalikan ErrNoTicketsAvailable pada request pertama", func(t *testing.T) {
		harness.Reset(t)
		// Inventory declared but never generated: the very first request
		// must already fail, not the one after the stock ran out.
		event := insertEvent(t, "No Rows", 3, time.Now().Add(-time.Minute))
		user := seedUser(t)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, services.ErrNoTicketsAvailable)
	})

	t.Run("masa jual belum dibuka mengembalikan ErrSaleNotOpen", func(t *testing.T) {
		harness.Reset(t)
		event := insertEvent(t, "Not Open", 5, time.Now().Add(2*time.Hour))
		insertTickets(t, event.ID, 5)
		user := seedUser(t)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, services.ErrSaleNotOpen)
	})

	t.Run("database menolak baris purchases kedua untuk ticket_id yang sama", func(t *testing.T) {
		event, user := seedSale(t, "Unique Guard", 2)
		ticket := lowestAvailable(t, event.ID)
		row := models.Purchase{TicketID: ticket.ID, UserID: user.ID, PriceCents: event.PriceCents}
		require.NoError(t, database.DB.Create(&row).Error)

		// The constraint is what is under test here, not the Go code: the
		// rule has to hold even when application code forgets it.
		err := database.DB.Create(&models.Purchase{
			TicketID: ticket.ID, UserID: user.ID, PriceCents: event.PriceCents,
		}).Error

		require.Error(t, err)
		assert.Contains(t, err.Error(), "purchases_ticket_id_key")
	})

	t.Run("INSERT purchases yang gagal mengembalikan tiket ke available", func(t *testing.T) {
		event, user := seedSale(t, "Rollback", 3)
		// Occupy the ticket the claim will pick, so its own INSERT trips
		// ticket_id UNIQUE *after* the UPDATE would have succeeded. A
		// non-transactional claim passes every other case here and fails
		// this one, leaving a seat sold that nobody owns.
		target := lowestAvailable(t, event.ID)
		require.NoError(t, database.DB.Create(&models.Purchase{
			TicketID: target.ID, UserID: user.ID, PriceCents: event.PriceCents,
		}).Error)
		before := countPurchases(t)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, models.TicketStatusAvailable, ticketStatus(t, target.ID))
		assert.Equal(t, before, countPurchases(t))
	})
}

func purchaseFlowEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("price_cents 9007199254740993 bertahan utuh melewati round trip database", func(t *testing.T) {
		// Above float64's exact-integer limit, so any float conversion
		// anywhere on the path turns this red.
		const exact int64 = 9007199254740993
		event, user := seedSale(t, "Precision", 2)
		require.NoError(t, database.DB.Model(&models.Event{}).
			Where("id = ?", event.ID).Update("price_cents", exact).Error)

		got, err := newService().Purchase(ctx, event.ID, user.ID)

		require.NoError(t, err)
		assert.Equal(t, exact, got.PriceCents)
		assert.Equal(t, exact, purchaseRow(t, got.TicketID).PriceCents)
	})

	t.Run("mengubah harga event setelah pembelian tidak mengubah baris purchases", func(t *testing.T) {
		event, user := seedSale(t, "Reprice", 2)
		got, err := newService().Purchase(ctx, event.ID, user.ID)
		require.NoError(t, err)

		require.NoError(t, database.DB.Model(&models.Event{}).
			Where("id = ?", event.ID).Update("price_cents", int64(1)).Error)

		// A snapshot, not a reference: repricing must not rewrite what
		// someone already paid.
		assert.Equal(t, event.PriceCents, purchaseRow(t, got.TicketID).PriceCents)
	})

	t.Run("tidak ada tiket sold tanpa baris purchases pasangannya", func(t *testing.T) {
		event, user := seedSale(t, "No Orphans", 5)
		service := newService()
		for i := 0; i < 3; i++ {
			_, err := service.Purchase(ctx, event.ID, user.ID)
			require.NoError(t, err)
		}

		var orphans int64
		require.NoError(t, database.DB.Raw(`
			SELECT COUNT(*) FROM tickets t
			LEFT JOIN purchases p ON p.ticket_id = t.id
			WHERE t.status = ? AND p.id IS NULL`,
			models.TicketStatusSold).Scan(&orphans).Error)

		// Inventory lost for good if this is ever non-zero, and no test
		// that only reads tickets.status would notice.
		assert.Equal(t, int64(0), orphans)
	})
}
