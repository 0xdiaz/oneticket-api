package orders_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	concurrentBuyers = 50
	inventory        = 20
)

// TestPurchaseTicket_Concurrent is the test that holds R8.
//
// It must fail against a naive read-then-write implementation. If it ever
// passes against one, the test is wrong, not the code.
func TestPurchaseTicket_Concurrent(t *testing.T) {
	t.Run("edge case", func(t *testing.T) {
		t.Run("concurrent buyers never oversell the event", func(t *testing.T) {
			harness.Reset(t)
			event := seedEvent(t, "Flash Sale Demo", inventory, 15000000)
			user := seedUser(t, "buyer@example.local")

			repo := repositories.NewOrderRepository()

			// Every goroutine blocks on the same gate and is released at once.
			// A race needs simultaneity, not volume: 50 sequential purchases
			// would never reproduce it.
			gate := make(chan struct{})
			var wg sync.WaitGroup
			var mu sync.Mutex
			var succeeded int
			var soldOut int
			var otherErr error

			for i := 0; i < concurrentBuyers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-gate

					_, _, err := repo.PurchaseTicket(event.ID, user.ID, 15000000)

					mu.Lock()
					defer mu.Unlock()
					switch {
					case err == nil:
						succeeded++
					case errors.Is(err, repositories.ErrNoTicketAvailable):
						soldOut++
					default:
						otherErr = err
					}
				}()
			}
			close(gate)
			wg.Wait()

			require.NoError(t, otherErr, "no purchase may fail for an unexpected reason")

			// Covers AE6: exactly the inventory sells, the rest are refused.
			assert.Equal(t, inventory, succeeded, "sold exactly the inventory")
			assert.Equal(t, concurrentBuyers-inventory, soldOut, "the rest were refused")

			// Covers AE6: the tickets table agrees.
			assert.Equal(t, int64(inventory), countTickets(t, event.ID, models.TicketStatusSold))
			assert.Equal(t, int64(0), countTickets(t, event.ID, models.TicketStatusAvailable))

			// Covers AE6: one order per ticket, and no ticket sold twice.
			var orderCount int64
			require.NoError(t, database.DB.Model(&models.Order{}).
				Where("event_id = ?", event.ID).Count(&orderCount).Error)
			assert.Equal(t, int64(inventory), orderCount)

			var distinctTickets int64
			require.NoError(t, database.DB.Model(&models.Order{}).
				Where("event_id = ?", event.ID).
				Distinct("ticket_id").Count(&distinctTickets).Error)
			assert.Equal(t, orderCount, distinctTickets,
				"every order holds a different ticket — no seat sold twice")
		})
	})
}
