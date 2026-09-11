// Fixtures and helpers for the checkout journeys.
//
// seedEventWithTickets is the one place E2E writes to the database directly:
// the e2e container is migrated but never seeded, so without it there is
// nothing to buy. Every assertion still goes over http.Client.
package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/internal/domain/models"
)

type purchase struct {
	ID          uint      `json:"id"`
	EventID     uint      `json:"event_id"`
	TicketID    uint      `json:"ticket_id"`
	TicketCode  string    `json:"ticket_code"`
	PriceCents  int64     `json:"price_cents"`
	PurchasedAt time.Time `json:"purchased_at"`
}

// seedEventWithTickets writes one event and n available tickets straight
// through database.DB.
//
// This is the one place in this package that touches the database directly,
// and it touches it only to build fixtures. No HTTP route creates an event,
// and the harness runs migrations without the dev seeders (those live in
// main.go behind APP_ENV=development), so the database this package starts
// against is empty and nothing is on sale. The rule E2E actually keeps is
// about the request and assertion path, and that stays pure: every claim
// below is made through call, over a real socket.
//
// tests/e2e never calls harness.Reset, so rows survive for the whole package.
// Each journey therefore seeds its own event and test order decides nothing.
func seedEventWithTickets(t *testing.T, name string, tickets int, saleStartsAt time.Time) *models.Event {
	t.Helper()

	ev := &models.Event{
		Name:         name,
		Venue:        "E2E Hall",
		StartsAt:     time.Now().Add(24 * time.Hour),
		SaleStartsAt: saleStartsAt,
		PriceCents:   15000000,
		TotalTickets: tickets,
	}
	if err := database.DB.Create(ev).Error; err != nil {
		t.Fatalf("seed event %q: %v", name, err)
	}

	rows := make([]models.Ticket, 0, tickets)
	for i := 1; i <= tickets; i++ {
		rows = append(rows, models.Ticket{
			EventID: ev.ID,
			Code:    fmt.Sprintf("A-%03d", i),
			Status:  models.TicketStatusAvailable,
		})
	}
	if tickets > 0 {
		if err := database.DB.Create(&rows).Error; err != nil {
			t.Fatalf("seed %d tickets for %q: %v", tickets, name, err)
		}
	}
	return ev
}

// signIn registers a fresh account and returns the access token the real login
// endpoint mints for it. A token a test hands itself would prove nothing.
func signIn(t *testing.T, prefix string) string {
	t.Helper()

	email := uniqueEmail(prefix)
	const password = "correct-horse-battery"

	status, env := call(t, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"name": "E2E Buyer", "email": email, "password": password,
	})
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("register answered %d, want 200 or 201: %s", status, env.Message)
	}

	status, env = call(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email": email, "password": password,
	})
	if status != http.StatusOK {
		t.Fatalf("login answered %d, want 200: %s", status, env.Message)
	}
	var issued tokens
	decode(t, env, &issued)
	if issued.AccessToken == "" {
		t.Fatal("login succeeded but returned an empty access_token")
	}
	return issued.AccessToken
}

// purchasePath takes any id, including the malformed ones, because a client
// can send anything in that segment and the server still owes it an answer.
func purchasePath(id any) string {
	return fmt.Sprintf("/api/v1/events/%v/purchase", id)
}

// readEvent fetches one event over HTTP, which is the only way this layer is
// allowed to observe availability.
func readEvent(t *testing.T, id uint) event {
	t.Helper()

	status, env := call(t, http.MethodGet, fmt.Sprintf("/api/v1/events/%d", id), "", nil)
	if status != http.StatusOK {
		t.Fatalf("GET /events/%d answered %d, want 200: %s", id, status, env.Message)
	}
	var got event
	decode(t, env, &got)
	return got
}

// TestJourney_VisitorBuysATicket is the checkout path end to end: arrive, sign
// up, sign in, see what is on sale, buy, and watch the inventory move.
//
// The last step is the one no unit test can make. Availability is counted from
// the tickets table on read, so a purchase that answered 201 without claiming
// a row would still look like a success to the buyer and show up only here.

var dbLeakMarkers = []string{"pgx", "int4", "OID", "SQL", "sql:"}

// assertNoInternalLeak re-encodes the envelope and scans it. call returns the
// decoded envelope rather than the bytes it read, and client_test.go is not
// this story's file; the re-encoded form still carries every field a client
// sees.
func assertNoInternalLeak(t *testing.T, env envelope) {
	t.Helper()

	body, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("re-encode response envelope: %v", err)
	}
	for _, marker := range dbLeakMarkers {
		if strings.Contains(string(body), marker) {
			t.Errorf("failure response leaked internal detail %q: %s", marker, body)
		}
	}
}

// TestJourney_PurchaseRejections covers every way a checkout can be refused.
// The status code is half of it; the other half is that a refusal says enough
// for a client to act on and nothing about the machinery behind it.
