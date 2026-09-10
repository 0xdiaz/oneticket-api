// Package harness boots the dependencies an integration test needs: a real
// Postgres in a container, migrated to the current schema, with the package
// level database handle wired to it.
//
// One container is shared per test package via TestMain. Starting a container
// per test would multiply a ~2.5s startup across the whole suite and blow the
// time budget.
//
// Usage, in each integration test package:
//
//	func TestMain(m *testing.M) { os.Exit(harness.RunMain(m)) }
//
// Tests then use the normal repositories; database.DB already points at the
// container. Call harness.Reset(t) to start a test from empty tables.
package harness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/adapters/database"
	"github.com/0xdiaz/oneticket-api/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// postgresImage is pinned so a test run never depends on what :latest happens
// to be, and so the image can be pre-pulled before a demo.
const postgresImage = "postgres:16-alpine"

// startupTimeout is generous on purpose: the first run on a cold machine may
// still be pulling the image.
const startupTimeout = 120 * time.Second

// dsn of the running container, set by RunMain.
var dsn string

// DSN returns the connection string of the container started for this package.
func DSN() string { return dsn }

// RunMain starts Postgres, applies migrations, wires database.DB, runs the
// package's tests, then tears the container down. It returns the exit code so
// the caller can hand it to os.Exit.
func RunMain(m *testing.M) int {
	ctx := context.Background()

	setTestConfig()

	container, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase("oneticket_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(startupTimeout)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness: start postgres: %v\n", err)
		return 1
	}
	// Terminate explicitly rather than with defer: os.Exit in the caller would
	// skip a deferred call, so the cleanup has to happen before returning.
	defer func() { _ = container.Terminate(ctx) }()

	dsn, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness: connection string: %v\n", err)
		return 1
	}

	// Master and replica point at the same instance; the split only matters in
	// production.
	if err := database.DbConnection(dsn, dsn); err != nil {
		fmt.Fprintf(os.Stderr, "harness: connect: %v\n", err)
		return 1
	}

	if err := applyMigrations(); err != nil {
		fmt.Fprintf(os.Stderr, "harness: migrate: %v\n", err)
		return 1
	}

	return m.Run()
}

// applyMigrations runs the versioned SQL against the container.
//
// migrations.Migrate() cannot be reused here: its path is a package constant
// relative to the repository root, and tests run with the working directory
// set to their own package. This resolves the same directory absolutely.
func applyMigrations() error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	sqlDir := filepath.Join(root, "internal", "adapters", "database", "migrations", "sql")
	if _, err := os.Stat(sqlDir); err != nil {
		return fmt.Errorf("migration directory not found at %s: %w", sqlDir, err)
	}

	sqlDB, err := database.GetDB().DB()
	if err != nil {
		return fmt.Errorf("underlying sql.DB: %w", err)
	}

	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://"+sqlDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("initialize migrator: %w", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// repoRoot walks up from the working directory until it finds go.mod.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// setTestConfig installs a configuration good enough for the layers under
// test, so nothing has to read a .env file.
func setTestConfig() {
	cfg := &config.Configuration{}
	cfg.Server.JWTSecret = "integration-test-jwt-secret-at-least-32-chars"
	cfg.Server.Debug = false
	cfg.Server.AccessTokenTTLMinutes = 15
	cfg.Server.RefreshTokenTTLDays = 7
	cfg.Server.Timezone = "UTC"
	cfg.Database.LogMode = false
	config.SetForTest(cfg)
}

// Reset empties every table the application owns, leaving the schema in place.
// Call it at the start of a test that needs to control the data completely.
//
// schema_migrations is deliberately untouched: truncating it would make the
// migrator think the database is empty.
func Reset(t *testing.T) {
	t.Helper()

	tables := []string{"tickets", "events", "refresh_tokens", "users", "examples"}
	for _, table := range tables {
		if err := database.DB.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatalf("harness: truncate %s: %v", table, err)
		}
	}
}
