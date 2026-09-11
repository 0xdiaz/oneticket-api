// Command smoke checks that a running server is alive and its critical paths
// work. It is the fastest question you can ask a deployment: is this thing
// actually up, and does the core flow respond?
//
// It proves wiring, not logic. Unit and integration tests already cover
// behaviour; this catches the failures they cannot see: a migration that did
// not run, a route that was dropped, a config that points at the wrong
// database, an auth guard that silently stopped guarding.
//
// Usage:
//
//	go run ./scripts/smoke
//	go run ./scripts/smoke -url http://staging.example.com
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type check struct {
	name string
	run  func(*client) error
}

func main() {
	url := flag.String("url", "http://localhost:8000", "base URL of the running server")
	email := flag.String("email", "admin@example.local", "seeded user to log in as")
	password := flag.String("password", "Password123!", "password for -email")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	flag.Parse()

	c := &client{
		base:     *url,
		email:    *email,
		password: *password,
		http:     &http.Client{Timeout: *timeout},
	}

	checks := []check{
		{"server menjawab", (*client).checkHealthReachable},
		{"database tersambung", (*client).checkHealthDatabase},
		{"metrics tersedia", (*client).checkMetrics},
		{"migrasi sudah jalan", (*client).checkSchemaMigrated},
		{"event ter-seed", (*client).checkEventsSeeded},
		{"login berhasil", (*client).checkLogin},
		{"route terlindungi menolak tanpa token", (*client).checkProtectedRejects},
		{"route terlindungi menerima token", (*client).checkProtectedAccepts},
		{"route tak dikenal menjawab 404", (*client).checkUnknownRoute},
		{"checkout menolak tanpa token", (*client).checkPurchaseRejectsAnonymous},
	}

	printHeader(c.base, len(checks))
	os.Exit(report(runAll(c, checks)))
}

func runAll(c *client, checks []check) []result {
	out := make([]result, 0, len(checks))
	for _, ch := range checks {
		start := time.Now()
		err := ch.run(c)
		out = append(out, result{name: ch.name, err: err, elapsed: time.Since(start)})
		printProgress(out[len(out)-1])
	}
	return out
}

type result struct {
	name    string
	err     error
	elapsed time.Duration
}

type client struct {
	base     string
	email    string
	password string
	http     *http.Client
	token    string
}

// get performs a GET and returns status plus body.
func (c *client) get(path, token string) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	if err != nil {
		return 0, nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, nil
}

// post performs a bodyless POST and returns status plus body.
func (c *client) post(path, token string) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPost, c.base+path, nil)
	if err != nil {
		return 0, nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, nil
}

func (c *client) checkHealthReachable() error {
	status, _, err := c.get("/health", "")
	if err != nil {
		return fmt.Errorf("tidak bisa menghubungi %s: %w", c.base, err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("/health menjawab %d, bukan 200", status)
	}
	return nil
}

func (c *client) checkHealthDatabase() error {
	_, body, err := c.get("/health", "")
	if err != nil {
		return err
	}
	var envelope struct {
		Data struct {
			Status string            `json:"status"`
			Checks map[string]string `json:"checks"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("respons /health bukan JSON yang dikenali: %w", err)
	}
	if envelope.Data.Checks["database"] != "ok" {
		return fmt.Errorf("database bukan ok, melainkan %q", envelope.Data.Checks["database"])
	}
	return nil
}

func (c *client) checkMetrics() error {
	status, _, err := c.get("/metrics", "")
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("/metrics menjawab %d", status)
	}
	return nil
}

// checkSchemaMigrated proves the migrations ran: the events endpoint can only
// answer 200 when its table exists.
func (c *client) checkSchemaMigrated() error {
	status, _, err := c.get("/api/v1/events", "")
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("/api/v1/events menjawab %d, tabelnya mungkin belum dimigrasi", status)
	}
	return nil
}

func (c *client) checkEventsSeeded() error {
	_, body, err := c.get("/api/v1/events", "")
	if err != nil {
		return err
	}
	var envelope struct {
		Data []struct {
			Name      string `json:"name"`
			Available int64  `json:"available_tickets"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if len(envelope.Data) == 0 {
		return fmt.Errorf("tidak ada event sama sekali, seeder belum jalan?")
	}
	return nil
}

func (c *client) checkLogin() error {
	payload := fmt.Sprintf(`{"email":%q,"password":%q}`, c.email, c.password)
	resp, err := c.http.Post(c.base+"/api/v1/auth/login", "application/json", strings.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login menjawab %d untuk %s", resp.StatusCode, c.email)
	}
	var envelope struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if envelope.Data.AccessToken == "" {
		return fmt.Errorf("login sukses tapi tidak mengembalikan access_token")
	}
	c.token = envelope.Data.AccessToken
	return nil
}

// checkProtectedRejects is the one that matters most. An auth guard that stops
// guarding still answers 200, so nothing else here would notice.
func (c *client) checkProtectedRejects() error {
	status, _, err := c.get("/api/v1/profile", "")
	if err != nil {
		return err
	}
	if status != http.StatusUnauthorized {
		return fmt.Errorf("/profile tanpa token menjawab %d, bukan 401. Guard-nya tidak menjaga", status)
	}
	return nil
}

func (c *client) checkProtectedAccepts() error {
	if c.token == "" {
		return fmt.Errorf("dilewati: login gagal lebih dulu")
	}
	status, _, err := c.get("/api/v1/profile", c.token)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("/profile dengan token valid menjawab %d", status)
	}
	return nil
}

func (c *client) checkUnknownRoute() error {
	status, _, err := c.get("/api/v1/tidak-ada-route-ini", "")
	if err != nil {
		return err
	}
	if status != http.StatusNotFound {
		return fmt.Errorf("route tak dikenal menjawab %d, bukan 404", status)
	}
	return nil
}

// checkPurchaseRejectsAnonymous proves checkout is alive without buying
// anything. A 401 means two things at once: the route is registered (an
// unregistered one answers 404) and the guard is attached to it (an
// unguarded one would reach the handler).
//
// It is deliberately the anonymous case. A check that actually bought a
// ticket would consume real inventory every time smoke ran, and a smoke
// check must not write.
func (c *client) checkPurchaseRejectsAnonymous() error {
	status, _, err := c.post("/api/v1/events/1/purchase", "")
	if err != nil {
		return err
	}
	if status == http.StatusNotFound {
		return fmt.Errorf("checkout menjawab 404: route-nya tidak terdaftar")
	}
	if status != http.StatusUnauthorized {
		return fmt.Errorf("checkout tanpa token menjawab %d, bukan 401. Guard-nya tidak menjaga", status)
	}
	return nil
}
