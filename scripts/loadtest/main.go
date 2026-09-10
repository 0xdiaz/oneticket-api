// Command loadtest fires concurrent ticket purchases at a running server and
// reports whether the event oversold.
//
// It is deliberately tolerant of the checkout contract, because the checkout
// endpoint is built live during the demo:
//
//   - Any 2xx counts as a successful purchase.
//   - A ticket code is extracted from the response when one is present, under
//     any of several likely field names, so duplicate allocation can be caught.
//   - Availability is read from GET /api/v1/events/{id} before and after, which
//     works no matter how checkout is implemented.
//
// The verdict is the point: it exits non-zero when the run oversold, so the
// result is a pass/fail on screen rather than a wall of numbers.
//
// Usage:
//
//	go run ./scripts/loadtest -n 500 -c 100
//	go run ./scripts/loadtest -n 500 -c 100 -path '/api/v1/events/{id}/purchase'
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type config struct {
	baseURL     string
	eventID     int
	requests    int
	concurrency int
	pathTmpl    string
	method      string
	body        string
	token       string
	email       string
	password    string
	timeout     time.Duration
}

func main() {
	cfg := parseFlags()

	client := &http.Client{Timeout: cfg.timeout}

	if cfg.token == "" && cfg.email != "" {
		token, err := login(client, cfg)
		if err != nil {
			fmt.Printf("  login gagal (%v) — lanjut tanpa token\n\n", err)
		} else {
			cfg.token = token
		}
	}

	before, err := availability(client, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tidak bisa membaca event %d: %v\n", cfg.eventID, err)
		fmt.Fprintf(os.Stderr, "server jalan di %s?\n", cfg.baseURL)
		os.Exit(2)
	}

	printHeader(cfg, before)

	res := run(client, cfg)

	after, err := availability(client, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tidak bisa membaca event setelah run: %v\n", err)
		os.Exit(2)
	}

	os.Exit(report(cfg, before, after, res))
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.baseURL, "url", "http://localhost:8000", "base URL of the running server")
	flag.IntVar(&cfg.eventID, "event", 1, "event id to hammer")
	flag.IntVar(&cfg.requests, "n", 500, "total purchase requests to send")
	flag.IntVar(&cfg.concurrency, "c", 100, "how many to keep in flight at once")
	flag.StringVar(&cfg.pathTmpl, "path", "/api/v1/events/{id}/purchase", "purchase path; {id} is replaced by -event")
	flag.StringVar(&cfg.method, "method", "POST", "HTTP method for the purchase call")
	flag.StringVar(&cfg.body, "body", "{}", "request body sent with each purchase")
	flag.StringVar(&cfg.token, "token", "", "Bearer token; when empty the tool logs in with -email/-password")
	flag.StringVar(&cfg.email, "email", "admin@example.local", "seeded user to log in as (empty to skip login)")
	flag.StringVar(&cfg.password, "password", "Password123!", "password for -email")
	flag.DurationVar(&cfg.timeout, "timeout", 30*time.Second, "per-request timeout")
	flag.Parse()
	return cfg
}

// eventState is the slice of GET /api/v1/events/{id} this tool cares about.
type eventState struct {
	Name      string `json:"name"`
	Total     int    `json:"total_tickets"`
	Available int64  `json:"available_tickets"`
}

func availability(client *http.Client, cfg config) (eventState, error) {
	url := fmt.Sprintf("%s/api/v1/events/%d", cfg.baseURL, cfg.eventID)
	resp, err := client.Get(url)
	if err != nil {
		return eventState{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return eventState{}, fmt.Errorf("GET %s -> %d", url, resp.StatusCode)
	}

	var envelope struct {
		Data eventState `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return eventState{}, err
	}
	return envelope.Data, nil
}

func login(client *http.Client, cfg config) (string, error) {
	payload := fmt.Sprintf(`{"email":%q,"password":%q}`, cfg.email, cfg.password)
	resp, err := client.Post(cfg.baseURL+"/api/v1/auth/login", "application/json", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login -> %d", resp.StatusCode)
	}

	var envelope struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return "", err
	}
	if envelope.Data.AccessToken == "" {
		return "", fmt.Errorf("login response had no access_token")
	}
	return envelope.Data.AccessToken, nil
}

// results collects what happened across all purchase attempts.
type results struct {
	mu       sync.Mutex
	statuses map[int]int    // HTTP status -> count
	codes    map[string]int // ticket code -> how many buyers got it
	failures map[string]int // transport error -> count
	elapsed  time.Duration
}

func run(client *http.Client, cfg config) *results {
	res := &results{
		statuses: map[int]int{},
		codes:    map[string]int{},
		failures: map[string]int{},
	}

	path := strings.ReplaceAll(cfg.pathTmpl, "{id}", strconv.Itoa(cfg.eventID))
	url := cfg.baseURL + path

	// Every worker blocks on the same gate so the requests land together
	// instead of trickling in; a race needs simultaneity, not just volume.
	gate := make(chan struct{})
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.concurrency)

	start := time.Now()
	for i := 0; i < cfg.requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			<-gate
			purchase(client, cfg, url, res)
		}()
	}
	close(gate)
	wg.Wait()
	res.elapsed = time.Since(start)

	return res
}

func purchase(client *http.Client, cfg config, url string, res *results) {
	req, err := http.NewRequestWithContext(context.Background(), cfg.method, url, bytes.NewReader([]byte(cfg.body)))
	if err != nil {
		res.recordFailure(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.token)
	}

	resp, err := client.Do(req)
	if err != nil {
		res.recordFailure(err)
		return
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	res.record(resp.StatusCode, ticketCode(payload))
}

func (r *results) record(status int, code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statuses[status]++
	if code != "" {
		r.codes[code]++
	}
}

func (r *results) recordFailure(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures[err.Error()]++
}

// succeeded counts every 2xx as one sold ticket.
func (r *results) succeeded() int {
	n := 0
	for status, count := range r.statuses {
		if status >= 200 && status < 300 {
			n += count
		}
	}
	return n
}

// duplicates returns the ticket codes handed to more than one buyer.
func (r *results) duplicates() map[string]int {
	dupes := map[string]int{}
	for code, count := range r.codes {
		if count > 1 {
			dupes[code] = count
		}
	}
	return dupes
}
