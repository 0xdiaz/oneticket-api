package purchases_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// attempts is the number of purchases fired at the event, and inFlight caps
// how many of them are actually in the database at once.
//
// The cap is not decoration. Nothing in the application bounds the connection
// pool, and postgres:16-alpine ships max_connections=100, so releasing all
// attempts simultaneously would end the run with "sorry, too many clients
// already" -- a failure that reads like a bug in the claim path and is not
// one. harness.RunMain caps the pool at 50 for the same reason, and
// scripts/loadtest bounds itself the same way.
const (
	attempts = 200
	inFlight = 50
	stock    = 100
)

// claimOutcome is what one goroutine's purchase attempt produced.
type claimOutcome struct {
	ticketID uint
	err      error
}

// TestPurchaseService_Purchase_Concurrent is the only test that puts the claim
// path under real contention.
//
// Everything it asserts is read back from the database rather than counted in
// memory: an in-memory tally only proves the goroutines agreed with each
// other, not that the rows are right.
//
// What it does NOT prove is that FOR UPDATE SKIP LOCKED is the mechanism in
// use. Every assertion below is satisfied identically by any correct
// serialisation -- a Go mutex, or a lock on the event row. The wall-clock
// duration reported at the end is the only signal that would move if the
// claim quietly became a queue, which is the same reason scripts/nplusone
// exists: a result can be correct and still be the wrong shape.
func TestPurchaseService_Purchase_Concurrent(t *testing.T) {
	t.Run("edge case", func(t *testing.T) {
		t.Run("200 percobaan terhadap 100 tiket menjual tepat 100", func(t *testing.T) {
			harness.Reset(t)
			event, user := seedSale(t, "Flash Sale Concurrent", stock)

			outcomes, elapsed := runConcurrentPurchases(t, event.ID, user.ID, attempts)

			var sold, soldOut int
			for _, out := range outcomes {
				switch {
				case out.err == nil:
					sold++
				case errors.Is(out.err, services.ErrNoTicketsAvailable):
					soldOut++
				default:
					// A connection-exhaustion or driver error here would
					// otherwise be miscounted as a concurrency defect.
					t.Errorf("unexpected error from a concurrent purchase: %v", out.err)
				}
			}

			assert.Equal(t, stock, sold, "should sell exactly the inventory")
			assert.Equal(t, attempts-stock, soldOut, "every loser should be told it is sold out")

			t.Logf("%d attempts, %d in flight, finished in %s", attempts, inFlight, elapsed)
		})

		t.Run("tidak ada tiket yang terjual dua kali", func(t *testing.T) {
			harness.Reset(t)
			event, user := seedSale(t, "No Double Sale", stock)

			runConcurrentPurchases(t, event.ID, user.ID, attempts)

			var rows, distinct int64
			require.NoError(t, database.DB.Model(&models.Purchase{}).Count(&rows).Error)
			require.NoError(t, database.DB.Model(&models.Purchase{}).
				Distinct("ticket_id").Count(&distinct).Error)

			assert.Equal(t, int64(stock), rows, "one purchases row per ticket sold")
			assert.Equal(t, rows, distinct, "no ticket_id appears twice")
		})

		t.Run("jumlah tiket sold sama dengan jumlah baris purchases", func(t *testing.T) {
			harness.Reset(t)
			event, user := seedSale(t, "No Orphans Under Load", stock)

			runConcurrentPurchases(t, event.ID, user.ID, attempts)

			var soldTickets, purchases int64
			require.NoError(t, database.DB.Model(&models.Ticket{}).
				Where("event_id = ? AND status = ?", event.ID, models.TicketStatusSold).
				Count(&soldTickets).Error)
			require.NoError(t, database.DB.Model(&models.Purchase{}).Count(&purchases).Error)

			assert.Equal(t, purchases, soldTickets,
				"a sold ticket with no purchases row is inventory lost for good")
		})

		t.Run("dua pembelian berbarengan oleh user yang sama menghasilkan tiket berbeda", func(t *testing.T) {
			harness.Reset(t)
			event, user := seedSale(t, "Same User Concurrent", 5)

			outcomes, _ := runConcurrentPurchases(t, event.ID, user.ID, 2)

			seen := make(map[uint]bool, 2)
			for _, out := range outcomes {
				require.NoError(t, out.err)
				require.False(t, seen[out.ticketID],
					"the same ticket was handed to both concurrent buyers")
				seen[out.ticketID] = true
			}
			assert.Len(t, seen, 2)
		})
	})
}

// runConcurrentPurchases fires n purchase attempts at one event, keeping at
// most inFlight of them in the database at a time, and returns what each one
// produced together with how long the whole wave took.
//
// A start channel releases the goroutines together so the first wave really
// does contend, rather than trickling in as each goroutine is scheduled.
func runConcurrentPurchases(t *testing.T, eventID, userID uint, n int) ([]claimOutcome, time.Duration) {
	t.Helper()

	service := newService()
	outcomes := make([]claimOutcome, n)
	sem := make(chan struct{}, inFlight)
	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(slot int) {
			defer wg.Done()
			<-start

			sem <- struct{}{}
			defer func() { <-sem }()

			response, err := service.Purchase(context.Background(), eventID, userID)
			out := claimOutcome{err: err}
			if err == nil && response != nil {
				out.ticketID = response.TicketID
			}
			outcomes[slot] = out
		}(i)
	}

	began := time.Now()
	close(start)
	wg.Wait()
	return outcomes, time.Since(began)
}
