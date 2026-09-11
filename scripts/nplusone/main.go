// Command nplusone measures how many database queries one call to the event
// list costs, as the number of events grows.
//
// A query count that grows with the row count is the N+1 signature. The
// verdict is the point: it exits non-zero when the count scales, so the result
// is a pass/fail on screen rather than a wall of numbers.
//
// It seeds its own events under a recognisable name prefix and removes them
// afterwards, so it can run against the development database without
// disturbing the demo data.
//
// Usage:
//
//	go run ./scripts/nplusone
//	go run ./scripts/nplusone -sizes 10,100,500
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
	"github.com/0xdiaz/oneticket-api/internal/domain/repositories"
	"github.com/0xdiaz/oneticket-api/pkg/config"
	"gorm.io/gorm"
)

// seedPrefix marks the rows this tool owns, so cleanup never touches demo data.
const seedPrefix = "nplusone-probe-"

// queries counts every statement GORM sends while a probe runs.
var queries int64

type result struct {
	events  int
	queries int64
	elapsed time.Duration
}

func main() {
	sizesFlag := flag.String("sizes", "10,100,200", "event counts to probe, comma separated")
	ticketsPer := flag.Int("tickets", 5, "tickets seeded per event")
	flag.Parse()

	sizes, err := parseSizes(*sizesFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := connect(); err != nil {
		fmt.Fprintf(os.Stderr, "tidak bisa connect ke database: %v\n", err)
		fmt.Fprintln(os.Stderr, "postgres jalan? cek .env dan docker compose")
		os.Exit(2)
	}
	installCounter()
	defer cleanup()
	cleanup() // clear leftovers from an interrupted run

	printHeader(sizes, *ticketsPer)

	service := services.NewEventService(
		repositories.NewEventRepository(),
		repositories.NewTicketRepository(),
	)

	results := make([]result, 0, len(sizes))
	for _, size := range sizes {
		cleanup()
		if err := seed(size, *ticketsPer); err != nil {
			fmt.Fprintf(os.Stderr, "seed gagal: %v\n", err)
			os.Exit(2)
		}
		results = append(results, probe(service, size))
	}

	os.Exit(report(results))
}

func parseSizes(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	sizes := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("ukuran tidak valid: %q", p)
		}
		sizes = append(sizes, n)
	}
	return sizes, nil
}

func connect() error {
	if err := config.SetupConfig(); err != nil {
		return err
	}
	// Silence GORM's statement log. This tool counts queries; printing all of
	// them would bury the verdict under hundreds of lines, exactly the thing
	// that must stay readable on a projector.
	if cfg := config.Get(); cfg != nil {
		cfg.Database.LogMode = false
		cfg.Server.Debug = false
	}
	master, replica := config.DbConfiguration()
	return database.DbConnection(master, replica)
}

// installCounter registers callbacks that tally every statement GORM issues.
// Counting here rather than in Postgres keeps the tool self-contained: no
// pg_stat_statements extension, no server configuration.
func installCounter() {
	count := func(*gorm.DB) { atomic.AddInt64(&queries, 1) }
	db := database.DB
	_ = db.Callback().Query().After("gorm:query").Register("nplusone:count_query", count)
	_ = db.Callback().Row().After("gorm:row").Register("nplusone:count_row", count)
	_ = db.Callback().Raw().After("gorm:raw").Register("nplusone:count_raw", count)
}

func seed(events, ticketsPer int) error {
	now := time.Now()
	for i := 0; i < events; i++ {
		event := models.Event{
			Name:         fmt.Sprintf("%s%d", seedPrefix, i),
			Venue:        "Probe Hall",
			StartsAt:     now.Add(24 * time.Hour),
			SaleStartsAt: now.Add(-time.Minute),
			PriceCents:   15000000,
			TotalTickets: ticketsPer,
		}
		if err := database.DB.Create(&event).Error; err != nil {
			return err
		}
		tickets := make([]models.Ticket, 0, ticketsPer)
		for j := 1; j <= ticketsPer; j++ {
			tickets = append(tickets, models.Ticket{
				EventID: event.ID,
				Code:    fmt.Sprintf("P-%04d", j),
				Status:  models.TicketStatusAvailable,
			})
		}
		if err := database.DB.Create(&tickets).Error; err != nil {
			return err
		}
	}
	return nil
}

// probe times one List call and counts the statements it costs.
func probe(service *services.EventService, size int) result {
	atomic.StoreInt64(&queries, 0)
	start := time.Now()
	_, err := service.List(context.Background())
	elapsed := time.Since(start)
	if err != nil {
		fmt.Fprintf(os.Stderr, "List gagal: %v\n", err)
		os.Exit(2)
	}
	return result{events: size, queries: atomic.LoadInt64(&queries), elapsed: elapsed}
}

func cleanup() {
	database.DB.Exec(
		"DELETE FROM tickets WHERE event_id IN (SELECT id FROM events WHERE name LIKE ?)",
		seedPrefix+"%")
	database.DB.Exec("DELETE FROM events WHERE name LIKE ?", seedPrefix+"%")
}
