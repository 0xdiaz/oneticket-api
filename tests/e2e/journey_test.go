package e2e_test

import (
	"net/http"
	"testing"
)

// tokens is the shape the auth endpoints answer with inside the envelope.
type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type profile struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type event struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	TotalTickets     int    `json:"total_tickets"`
	AvailableTickets int    `json:"available_tickets"`
	PriceCents       int64  `json:"price_cents"`
}

// TestJourney_NewVisitorBrowsesAndSignsIn walks the path a real user walks, in
// order, on one connection: arrive, look at what is on sale, sign up, sign in,
// read their own profile.
//
// It is written as one test rather than five because the value is in the
// sequence. A token that works only because a test handed it to itself proves
// nothing; this one comes out of the login endpoint and goes back in through
// the Authorization header, which is exactly the trip it makes in production.
func TestJourney_NewVisitorBrowsesAndSignsIn(t *testing.T) {
	// Browsing is public. This has to work before any account exists.
	status, env := call(t, http.MethodGet, "/api/v1/events", "", nil)
	if status != http.StatusOK {
		t.Fatalf("GET /events answered %d, want 200: %s", status, env.Message)
	}
	var events []event
	decode(t, env, &events)

	email := uniqueEmail("visitor")
	const password = "correct-horse-battery"

	status, env = call(t, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"name":     "E2E Visitor",
		"email":    email,
		"password": password,
	})
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("register answered %d, want 200 or 201: %s", status, env.Message)
	}

	status, env = call(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if status != http.StatusOK {
		t.Fatalf("login answered %d, want 200: %s", status, env.Message)
	}
	var issued tokens
	decode(t, env, &issued)
	if issued.AccessToken == "" {
		t.Fatal("login succeeded but returned an empty access_token")
	}

	// The token from login has to be accepted by the guard on a different
	// route. This is the wiring E2E exists to check.
	status, env = call(t, http.MethodGet, "/api/v1/profile", issued.AccessToken, nil)
	if status != http.StatusOK {
		t.Fatalf("GET /profile with a freshly issued token answered %d, want 200: %s",
			status, env.Message)
	}
	var me profile
	decode(t, env, &me)
	if me.Email != email {
		t.Errorf("profile returned email %q, want %q", me.Email, email)
	}
}

// TestJourney_GuardRejectsUnauthenticated is the negative half, and the half
// that matters more. A guard that has stopped guarding still answers 200, so
// only a test that expects a refusal can catch it.
func TestJourney_GuardRejectsUnauthenticated(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		token  string
	}{
		{"tanpa header sama sekali", http.MethodGet, "/api/v1/profile", ""},
		{"token sampah", http.MethodGet, "/api/v1/profile", "not-a-jwt"},
		{"token dipotong", http.MethodGet, "/api/v1/profile", "eyJhbGciOiJIUzI1NiJ9.truncated"},
		{"logout-all tanpa token", http.MethodPost, "/api/v1/logout-all", ""},
		{"beli tanpa token", http.MethodPost, "/api/v1/events/1/purchase", ""},
		{"beli dengan token sampah", http.MethodPost, "/api/v1/events/1/purchase", "not-a-jwt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, env := call(t, tc.method, tc.path, tc.token, nil)
			if status != http.StatusUnauthorized {
				t.Fatalf("%s %s answered %d, want 401: %s",
					tc.method, tc.path, status, env.Message)
			}
			if env.Success {
				t.Error("a rejected request came back with success=true")
			}
		})
	}
}

// TestJourney_UnknownRouteIsHandled checks the fallback. Gin answers an
// unregistered path with its own 404 body unless NoRoute is set, which would
// hand a client a response shaped differently from every other response.
func TestJourney_UnknownRouteIsHandled(t *testing.T) {
	status, env := call(t, http.MethodGet, "/api/v1/there-is-no-such-thing", "", nil)
	if status != http.StatusNotFound {
		t.Fatalf("unknown route answered %d, want 404", status)
	}
	if env.Success {
		t.Error("404 came back with success=true")
	}
}

// purchase is the shape POST /events/{id}/purchase answers with inside the
// envelope. It mirrors dto.PurchaseResponse by field name only: what is under
// test is the JSON a client reads, so the type is declared here rather than
// imported.
