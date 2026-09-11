# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Go 1.25 + Gin REST API (`github.com/0xdiaz/oneticket-api`), PostgreSQL via GORM, JWT auth with
refresh token rotation. Derived from `0xdiaz/gin-boilerplate`.

The domain is **flash sale ticketing**: an event has a fixed number of tickets, one row per seat,
and a lot of people press buy at the same second. Today only the read path exists,
`GET /api/v1/events` and `/events/:id`. **Checkout is deliberately absent**; it is built live
during a sharing session on agentic development (`docs/DEMO_RUNSHEET.md`). Do not implement it
casually, see "The demo constraint" below.

## The demo constraint

This repository doubles as the working codebase for a live demo, and two things are missing **on
purpose**:

1. **No checkout endpoint.** A reference implementation lives on branch `demo/work-ready`; the plan
   for it is on `demo/plan-ready`. `main` must stay without it.
2. **`EventService.List` has an N+1.** It counts availability once per event. This is the bug the
   demo diagnoses on stage. `go run ./scripts/nplusone` reports `200 event -> 202 query`.
3. **`GET /api/v1/events/{id}` answers 500 on an out-of-range id**, and leaks the driver's own
   message while doing it (`unable to encode ... into binary format for int4 (OID 23)`). The id is
   parsed as `uint64` and handed to an `int4` column. Both `scripts/apitest` and `scripts/security`
   find it independently, which is the demo's point: two tools that know nothing about each other
   stop at the same line.
4. **No `X-Content-Type-Options` header.** A configuration gap that no Go test in this repo covers,
   because no handler is wrong.

If a task asks you to add checkout, fix the N+1, or fix either scanner finding, do it, but say
plainly that it removes a demo beat, and prefer a branch over `main`.

## Security requirements

**IMPORTANT**: Follow these on every implementation task. They supplement your engineering
instructions; do not restate or role-play around them.

Before implementing any feature touching **user-controlled data** (headers, filenames, IDs,
third-party payloads), **trace where it is stored and every place it is used, then secure each sink
in context**: parameterized queries for SQL, allowlist validation for structured fields.
**Store raw; encode on output.** Bind only allowed fields (no mass assignment); return only needed
data.

**Never** blacklist or strip characters from free text (`<`, `'`, `;`), sanitize on write, or block
valid input like `O'Brien`. **Prefer framework-native, structural controls** that make the bug
impossible over string surgery. **Match strength to exposure × impact**, "low risk, standard
handling" is valid. **No speculative controls; no secrets in code, logs, or URLs.**

**Fail closed once:** on a failed check, reject (4xx) + log server-side; **never degrade silently,
add fallback logic** (`sanitize(x) || x`)**, or duplicate one protection across layers.** Verify per
feature, flagging only what applies: object-level authz (IDOR), rate limits on abuse-prone flows,
and no verbose errors or sensitive data leaked to clients.

**This is not hypothetical here.** `POST /auth/forgot-password` used to return the raw password
reset token in the response body, in every environment, because the mailer was wired `nil`
unconditionally and no `EmailSender` implementation existed. Anyone who knew a victim's email could
take the account in two requests. Worse, the test suite **asserted** the vulnerable behaviour
(`assert.Len(t, token, 64)`), so fixing it turned the suite red and looked like the fixer's mistake.
It is fixed now. The endpoint fails closed with 503 outside development, but treat "the test
passes" as weaker evidence than you would like.

## Commands

```bash
make run                 # go run main.go (needs .env, see Runtime prerequisites)
make dev                 # full docker stack: postgres + pgadmin + API with Air live reload
make test                # unit tests
make test-integration    # integration tests (Testcontainers, no setup needed)
make test-all            # unit + integration with -race
make loadtest            # reset demo event, then the flash sale oversell probe
go run ./scripts/nplusone            # query-count probe (N+1 detector)
go test ./tests/unit/services -run TestEventService -v   # single test
make migrate-create NAME=add_something
make migrate-down        # rollback last migration (needs the golang-migrate CLI)
```

**There is no linter config and no CI.** No `.golangci.yml`, no `.github/workflows`. The gate is
what you run locally:

```bash
gofmt -w . && go vet ./... && go build ./... && go test ./tests/... -race
```

Because nothing enforces this, a rule stated only in prose is a rule that will drift. When you add
a convention worth keeping, prefer a mechanism that fails loudly, a compile-time assertion, a
test, a database constraint, over a sentence in a document.

**`make test-coverage-check` is broken in two independent ways. Do not trust it.**

1. It runs `go test -coverprofile` across three package patterns, which needs the `covdata` tool.
   The toolchain in use here (`golang.org/toolchain@v0.0.1-go1.25.3`) does not ship it:
   `go: no such tool "covdata"`.
2. Even fixed, it would report `0.0%`. Tests live in `tests/unit/...`, separate packages from the
   code they exercise, so without `-coverpkg` the profile measures the test packages themselves,
   which have no statements.

Real coverage is **31.3%**, obtained with:

```bash
go test -coverpkg=./internal/...,./pkg/... -coverprofile=coverage.out ./tests/unit/...
go tool cover -func=coverage.out | tail -1
```

The 70% floor in the Makefile therefore guards nothing: it measures the wrong thing and cannot
execute. Fix the measurement before anyone quotes the number.

**`gofmt -w` only, never `goimports`.** Module imports are grouped above third-party here;
`goimports` resorts them and inflates every diff.

## Runtime prerequisites

- **Config comes from the `.env` FILE, not the environment.** `config.SetupConfig()` calls
  `viper.SetConfigFile(".env")` and does **not** call `viper.AutomaticEnv()`. Real environment
  variables are ignored, which has one consequence worth internalising: **`environment:` entries in
  docker-compose do not affect application config.** The container reads the same mounted `.env`
  the host does.

  That is why `MASTER_DB_HOST` in `.env` decides which run mode you are in:

  | Mode | `.env` | How to run |
  |---|---|---|
  | Postgres in Docker, API on host | `localhost` / `5436` | `docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db` then `go run main.go` |
  | Everything in Docker | `postgres_db` / `5432` | `make dev` |

  `ValidateConfig()` enforces eight required keys, `JWT_SECRET` ≥32 chars and free of
  `CHANGE-THIS`/`your-jwt-secret`, and well-formed `CORS_ALLOWED_ORIGINS` (required in production).
  Copying `.env.example` verbatim fails that check by design.

- **`--env-file` is mandatory for every compose command.** The compose files live in `.docker/`, so
  without it `${MASTER_DB_*}` interpolates to empty strings and compose fails with the unhelpful
  `invalid proto:`. Every Makefile target already passes it.

- **Container names and ports are project-scoped** (`oneticket_pg_db`, host port `5436`;
  pgadmin `5051`). The upstream boilerplate hardcodes `gin_pg_db` on `5434`, so both can run at
  once. Do not rename them back.

- **`air` is pinned to `v1.67.1` in `.docker/Dockerfile-dev`.** From v1.67.2 it requires Go ≥1.26
  while the base image is `golang:1.25-alpine`, so `@latest` breaks the dev image with no change to
  this repository. Raise the pin only together with the base image.

- Startup order in `main.go`: config → `time.Local` → metrics → DB connect → **SQL migrations
  (fatal on error)** → dev-only seeders → engine → serve with graceful shutdown. Migrations run on
  every boot and a failure is fatal, so the server never serves against a half-migrated schema.

- **Migrations are read from disk, not embedded.** `migrationsPath` is the relative constant
  `file://internal/adapters/database/migrations/sql`, so `migrations.Migrate()` only works with the
  repository root as the working directory. Anything running from elsewhere, a test binary, a
  script. Must resolve the directory itself; `tests/integration/harness` walks up to `go.mod` and
  builds an absolute path for exactly this reason.

- **SQL is the only source of schema truth.** There is no AutoMigrate. Every schema change is a new
  `{seq}_{name}.up.sql` / `.down.sql` pair in `internal/adapters/database/migrations/sql/`, and the
  GORM struct tags must match the DDL by hand. Never edit an applied migration.

- Dev seeds two users (`admin@example.local` / `user@example.local`, password `Password123!`) plus
  one event, "Flash Sale Demo", with 100 tickets. Seeders run only when `APP_ENV=development` and
  are idempotent.

- Swagger UI is served at `/swagger/` only when `DEBUG=true`. The spec in `api/openapi.yaml` is
  **hand-written** and embedded via `api/spec.go`, there are no swag annotations, so a route change
  must be mirrored there by hand. It currently matches the registered routes exactly (13 paths);
  that number is worth preserving.

## Architecture

**Package-by-layer**, not package-by-feature. `docs/MODULE_GUIDE.md` is the authoritative spec.
A feature is one file in each layer directory, tied together by a shared filename prefix.

```
main.go                          config → db → migrate → seed → serve
api/                             hand-written openapi.yaml, embedded via spec.go
internal/adapters/database/      connection, migrations/ (runner + sql/), seeders/
internal/app/controllers/        HTTP layer, <name>_controller.go
internal/app/dto/                request/response types
internal/app/middlewares/        auth, cors, metrics, rate_limit, request_id, request_log
internal/app/routers/            router.go (engine), index.go (the only wiring),
                                 <name>_routes.go per feature, swagger.go
internal/app/services/           business logic, <name>_service.go; auth/ is its own package
internal/domain/models/          GORM structs with TableName()
internal/domain/repositories/    interface + unexported impl + New*Repository()
pkg/                             config, logger, metrics, types, utils, never imports internal/
tests/unit/<layer>/              package <layer>_test, no database
tests/integration/harness/       one Postgres container per package via Testcontainers
tests/mocks/                     shared in-memory fakes
scripts/loadtest/                flash sale oversell probe
scripts/nplusone/                query-count probe (N+1 detector)
docs/                            standards, patterns, demo runsheet and prompts
```

The reference slice is **`event`**: `event_model.go` → `event_repo.go` → `event_dto.go` →
`event_service.go` → `event_controller.go` → `event_routes.go` → `event_repo_mock.go` →
`event_service_test.go`. Copy its shape. The older `example` slice uses a package-level function
for its repository instead of the interface pattern, follow `event`, not `example`.

Wiring lives only in `internal/app/routers/index.go`. It constructs repositories, injects them into
services, and calls each feature's `Register<Name>Routes`. Adding a feature touches that file once.

`/health` and `/metrics` mount at the **root** via `RegisterHealthRoutes(route)`, deliberately
outside `/api/v1` so they skip the rate limiter, a throttled `/health` fails the container
healthcheck exactly when it matters.

## Conventions that matter here

- **Responses**: never `c.JSON` in a controller. Use `pkg/utils` (`utils.Ok`, `utils.Created`,
  `utils.Conflict`, `utils.RespondWithAPIError`, …). Every response shares
  `{success, message, data, errors}`, a stable contract, see `docs/CONTRACTS.md`.

- **Domain errors → HTTP**: services return sentinel errors (`ErrEventNotFound`,
  `ErrInvalidCredentials`, …); the controller maps them with `errors.Is` and falls back to
  `utils.InternalServerError`. Do not leak status decisions into services.

- **Repositories return `(nil, nil)` for a missing row**, translating `gorm.ErrRecordNotFound`
  themselves. The service decides whether absence is an error. This is a real trade-off and the
  cost is worth naming: absence and an outage are distinguished by *discipline* rather than by
  error value, so a repository that forgets the translation reports an outage as "not found". If
  you change one repository to a sentinel, change them all, a codebase where half do each is worse
  than either.

- **Repositories are the only layer that touches `database.DB`.** It is a package-level handle in
  `internal/adapters/database`, not an injected `*gorm.DB`. The testability seam is the
  **interface**, not the connection: services depend on `repositories.XRepository`, so
  `tests/mocks/` substitutes without a database.

- **Money is `int64` in the smallest unit** (`price_cents`), never a float, at every layer
  including JSON. `TestEventService_Get/edge_case/price_survives_the_database_round_trip`
  uses `9007199254740993`. Above `float64`'s exact-integer limit, so any float conversion
  anywhere on the path turns it red.

- **Tracing**: every controller and service method wraps its body in
  `ctx, start := logger.LogStart(ctx, "EventService.Get")` and calls `logger.LogFinish` before
  *every* return. Span name is `<Type>.<Method>`, no module prefix, unlike some sibling projects.
  Repositories trace nothing. `request_id` flows through `context.Context`.

- **Tests live in `tests/`, not beside the code.** Unit tests in `tests/unit/<layer>/` as
  `package <layer>_test` (black box, driving the exported surface); shared fakes in `tests/mocks/`;
  anything needing a database in `tests/integration/`. Naming is
  `describe method → describe positive/negative/edge case → test`.

- **Every fake carries a compile-time assertion**:
  `var _ repositories.EventRepository = (*MockEventRepository)(nil)`. This is the repository's only
  real enforcement: with no linter and no CI, that line is what turns an interface change into a
  compile error instead of a silently stale fake. It has already earned its place, adding a batch
  count method to `TicketRepository` failed the build at the mock rather than passing with a rotten
  test.

- **Integration tests use the Testcontainers harness**, never a DSN or a skip.
  `func TestMain(m *testing.M) { os.Exit(harness.RunMain(m)) }` starts one Postgres per package,
  migrates it, and points `database.DB` at it; `harness.Reset(t)` truncates between tests. One
  container per **package**, never per test, startup is ~2.5s and would multiply. Docker is the
  only prerequisite; it works with OrbStack unmodified. The two older tests reading
  `TEST_DB_MASTER_DSN` and skipping are legacy; do not copy that pattern.

  The harness terminates its container inside `RunMain` rather than via `defer` in the caller,
  because `os.Exit` in `TestMain` skips deferred calls and would leak a container per package.
  `RunMain` also takes optional `before` hooks that run after migrations and before any test; the
  e2e package uses one to stand up its HTTP server, which cannot be built earlier because its
  handlers resolve repositories that need `database.DB` already wired.

- **E2E tests live in `tests/e2e/` and are a different layer from integration**, not a thicker
  version of it. Integration calls services directly and proves a service and its repository agree
  about the database. E2E goes over a real socket through the whole middleware chain, with a token
  minted by the real login endpoint, and proves the pieces are wired together at all. Use
  `routers.SetupRoute()`, never a router assembled inside the test: a test that registers its own
  routes proves nothing about the ones the application registers.

  This layer owns a bug class no other layer sees. Comment out the `AuthMiddleware` line in
  `internal/app/routers/index.go` and all six unit and integration packages stay green while
  `tests/e2e` goes red. A handler can be correct and unreachable; a guard can be correct and
  attached to the wrong group.

- **Constraints belong in the migration, not only in Go.** `CHECK (status IN ('available','sold'))`
  and `UNIQUE (event_id, code)` are enforced by the database, so an application bug cannot write a
  row that violates them. Keep the Go constants (`models.TicketStatusAvailable`) in sync with the
  CHECK by hand. This is load-bearing: an integration test asserting an event with zero tickets
  failed because `CHECK (total_tickets > 0)` refused the row, the test was wrong, not the code,
  which is a correction only a real database can make.

- Hard limits from `docs/00_AI_CRITICAL_RULES.md`: file ≤ 300 lines, function ≤ 100 lines.

- Any route change must be mirrored by hand in `api/openapi.yaml`.

## Measuring, not just testing

Five tools live in `scripts/` because five classes of defect pass every test. None of them is a
`go test`, and none belongs behind `make` during a demo: their exit code is the verdict, so make
appends `make: *** Error 1` right after the result.

The exit code only survives a compiled binary. `go run` prints `exit status 3` as text and exits 1
itself, so script anything that branches on the code with `go build -o probe ./scripts/<tool>`
first. The two shell wrappers (`apitest`, `security`) pass their codes through unchanged.

| Tool | Question it answers | Needs a live server |
|---|---|---|
| `scripts/loadtest` | Does concurrency corrupt the inventory | yes |
| `scripts/nplusone` | What does one request cost, not whether it is right | no |
| `scripts/smoke` | Is the thing that shipped actually alive | yes |
| `scripts/apitest` | Do the answers match the API's own spec | yes |
| `scripts/security` | Is anything leaking outside the handlers | yes |

Each is described below.

**`scripts/loadtest`**, concurrent purchases against a fixed inventory. Reports whether the event
oversold and exits non-zero if it did. Verified against a naive implementation (300 sold from 100,
27 tickets sold twice), against a locked one (100 from 100), and against no endpoint at all.

**`scripts/nplusone`**. Counts the database statements one `GET /api/v1/events` costs as the row
count grows. `200 event -> 202 query` is the N+1 signature; the fixed shape is a flat `2`.

Neither is a test, and that is the point: **N+1 passes the entire suite.** The result is correct,
just expensive, no assertion fails, nothing is red. When you change a read path that fans out over
rows, or a write path under concurrency, run the probe rather than trusting green.

**`scripts/smoke`** answers a different question from every test above: not whether the code is
correct, but whether the thing that is running is actually running. Nine cheap read-only checks
against a live server. The one that matters most is that `/api/v1/profile` **refuses** a request
with no token, because a guard that has stopped guarding still answers 200 and nothing else would
notice.

**`scripts/apitest`** points Schemathesis at `api/openapi.yaml`. It writes no test code: every case
comes from the contract already in the repo, so the suite grows on its own whenever the spec gains
a field. That makes keeping the spec honest a testing job, not a documentation job. Read-only
endpoints by default; `--all` includes the ones that write, which needs a throwaway database.

**`scripts/security`** runs OWASP ZAP against the same spec. It covers what lives outside the
handlers: response headers, information leaked by the error format, transport settings. Passive by
default, `--active` sends real attack payloads and belongs nowhere near data you care about.

**Regression is a rule, not a folder.** There is no `tests/regression/`. Every bug fix starts from
one test that is red first, and that test then stays in whichever layer it was born in. A test born
from a fixed bug is a regression test wherever it sits.

## Committing

Small commits at natural seams, as the work proceeds, not one commit at the end. For a typical
feature that is roughly:

1. `feat(x): migration NNNNNN + model`. The SQL pair and the GORM struct together; neither is
   meaningful alone
2. `feat(x): repository + fake`. The interface, its implementation, and the mock that asserts
   against it
3. `test(x): <what the tests pin>`, written before the implementation when the work is
   behaviour-bearing
4. `feat(x): service`
5. `feat(x): routes, controller, openapi`. The contract surface, which changes together or not
   at all

Stage explicitly by path; never `git add .`. Each commit should build.

## Which docs to trust

All of `docs/` was realigned with the code. The doc set previously described a package-by-feature
layout (`internal/modules/`, `internal/bootstrap/`, co-located tests) that never existed here, and
`MODULE_GUIDE.md` declared itself authoritative while showing this repository's actual pattern as a
❌ WRONG example.

- **Authoritative:** `docs/MODULE_GUIDE.md` (layout, how to add a feature),
  `docs/00_AI_CRITICAL_RULES.md` (hard limits, read first), `docs/CONTRACTS.md` (response envelope),
  `docs/MIGRATIONS.md` (schema workflow).
- **Reference:** `docs/CODING_STANDARDS.md`, `docs/DESIGN_PATTERNS.md`, `docs/AI_AGENT_RULES.md`,
  `docs/AI_QUICK_REFERENCE.md`. Long, and now accurate.
- **Marked NOT IMPLEMENTED:** the OpenTelemetry section of `docs/OBSERVABILITY.md`. It describes
  `pkg/observability`, `internal/clients` and `pkg/pii`, none of which exist and none of which are
  in `go.mod`. It is kept as a design, not a description.
- **Point-in-time:** `docs/CONTROLLER_COMPLIANCE_AUDIT.md`, `docs/SERVICE_COMPLIANCE_AUDIT.md`.
  They audit the current layout and their findings still apply, but re-run them after significant
  change rather than assuming they are current.
- **Demo-only:** `docs/DEMO_RUNSHEET.md` and `docs/demo/` (runsheet, paste-ready prompts, the
  answer sheet from the planning dry run). Read them before touching anything the demo constraint
  above protects.

  `docs/brainstorm/` and `docs/plans/` are **not on `main` on purpose**, the demo produces them
  live. Reference copies from the dry run are on branch `demo/plan-ready`, which is also the
  parachute if the live planning segment overruns.

Known rough edges deliberately left alone, each wanting its own change:

- `EventService.List` is N+1 (protected by the demo constraint).
- Config ignores the process environment, so container config must go through the mounted `.env`.
- The coverage target is broken twice over, as described under Commands.
- `internal/domain/repositories/example_repo.go` is a package-level function rather than the
  interface pattern every other repository follows.
