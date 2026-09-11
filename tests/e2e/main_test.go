// Package e2e drives the application the way a client does: over a real
// socket, through every middleware, against a real Postgres.
//
// It is a different layer from tests/integration, which calls services
// directly. Integration proves a service and its repository agree about the
// database. E2E proves the pieces are actually wired together: that the route
// is registered, that the middleware chain lets the request through in the
// right order, that a token minted by the login endpoint is accepted by the
// auth guard, and that the JSON on the wire has the field names a client
// reads. Nothing here calls an internal function. Every assertion goes through
// http.Client.
//
// A whole class of bug lives only at this layer. A handler can be correct and
// still be unreachable because nobody registered its route; a guard can be
// correct and still be bypassed because it was attached to the wrong group.
// No unit or integration test in this repo would notice either one.
package e2e_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/0xdiaz/oneticket-api/internal/app/routers"
	"github.com/0xdiaz/oneticket-api/pkg/metrics"
	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
)

// server is the running application under test, shared by every test in the
// package. It is built by startServer, after the harness has migrated the
// container and wired database.DB.
var server *httptest.Server

func TestMain(m *testing.M) {
	code := harness.RunMain(m, startServer)
	if server != nil {
		server.Close()
	}
	os.Exit(code)
}

// startServer builds the same router main.go builds and serves it on a real
// port. SetupRoute is used deliberately instead of hand-assembling routes: a
// test that registers its own routes proves nothing about the ones the
// application registers.
func startServer() error {
	metrics.Init()
	server = httptest.NewServer(routers.SetupRoute())
	return nil
}
