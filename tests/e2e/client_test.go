package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// envelope is the wrapper every endpoint in this API answers with. Decoding
// into it, rather than into the payload directly, is part of what E2E checks:
// a handler that returns a bare object instead of the wrapper is a contract
// break a client would feel and no internal test would see.
type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Errors  json.RawMessage `json:"errors"`
}

// call sends a request to the running server and returns the status and the
// decoded envelope. token may be empty, in which case no Authorization header
// is sent at all, which is the case the auth guard has to reject.
func call(t *testing.T, method, path, token string, body any) (int, envelope) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, server.URL+path, reader)
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s body: %v", method, path, err)
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("%s %s answered %d with a body that is not the standard envelope: %s",
			method, path, resp.StatusCode, raw)
	}
	return resp.StatusCode, env
}

// decode unpacks the data field of an envelope into target.
func decode(t *testing.T, env envelope, target any) {
	t.Helper()
	if err := json.Unmarshal(env.Data, target); err != nil {
		t.Fatalf("decode data field: %v (raw: %s)", err, env.Data)
	}
}

// uniqueEmail keeps tests independent of each other and of whatever the
// previous run left behind.
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@e2e.test", prefix, time.Now().UnixNano())
}
