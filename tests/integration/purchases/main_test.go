package purchases_test

import (
	"os"
	"testing"

	"github.com/0xdiaz/oneticket-api/tests/integration/harness"
)

// TestMain boots one Postgres container for this whole package.
func TestMain(m *testing.M) {
	os.Exit(harness.RunMain(m))
}
