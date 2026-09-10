# ⚠️ AI CRITICAL RULES - READ THIS FIRST

> **For AI Agents:** Read this BEFORE touching ANY code. These are NON-NEGOTIABLE rules.

> 🧭 **Structure:** Code is organized **by technical layer** —
> `internal/app/controllers` → `internal/app/services` → `internal/domain/repositories` →
> `internal/domain/models`, wired in `internal/app/routers/index.go`.
> See **[MODULE_GUIDE.md](./MODULE_GUIDE.md)** (the source of truth for the layout).
> The rules below — struct + constructor injection, response utilities, error-as-value +
> APIError mapping, file-size limits — are the non-negotiables that apply *inside* that layout.

> 🔐 **Security posture:** Every change must align with **OWASP Top 10:2025** and **OWASP WSTG**.
> Before touching auth, input handling, PII, money movement, cryptography, or external I/O —
> review the applicable OWASP categories. When this service handles sensitive personal data or
> financial operations, that obligation is Tier 0.

---

## 🚨 TIER 0: ABSOLUTE RULES (NEVER VIOLATE)

### 1. Architecture Pattern (MANDATORY)

```go
❌ WRONG - Standalone functions on globals:
func Register(c *gin.Context) { }
func Login(c *gin.Context) { }

✅ CORRECT - Struct-based with constructor injection:
// internal/app/controllers/event_controller.go
type EventController struct {
    service *services.EventService
}

func NewEventController(service *services.EventService) *EventController {
    return &EventController{service: service}
}

func (ctrl *EventController) List(c *gin.Context) { }
func (ctrl *EventController) Get(c *gin.Context) { }
```

**Rule:** Controllers and Services MUST be structs with methods, wired via `New*` constructors.
NO standalone functions operating on package globals. Business logic lives on a Service method,
never in a controller.

**Repositories** declare an exported interface plus an unexported implementation:

```go
// internal/domain/repositories/event_repo.go
type EventRepository interface {
    List() ([]*models.Event, error)
    GetByID(id uint) (*models.Event, error)
}

type eventRepo struct{}

func NewEventRepository() EventRepository { return &eventRepo{} }
```

Services depend on that **interface**, never on the concrete type — that is what makes them
testable with the fakes in `tests/mocks/`.

### 2. Response Format (MANDATORY)

```go
❌ WRONG - Direct gin.H:
c.JSON(200, gin.H{"status": 200, "data": user})

✅ CORRECT - Use Response Utilities:
utils.Ok(c, user, "User retrieved successfully")
utils.Created(c, user, "User created successfully")
utils.BadRequest(c, err, "Invalid input")
utils.Unauthorized(c, err, "Invalid credentials")
```

**Standard Format:**
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {...},
  "errors": null
}
```

**Rule:** ALL responses MUST use `pkg/utils/response.go` utilities. NO direct `c.JSON()`.

### 3. Test Location (MANDATORY)

```
✅ CORRECT - In the tests/ tree, black box, shared fake (no DB):
tests/unit/services/event_service_test.go     (package services_test)
tests/mocks/event_repo_mock.go                (package mocks)

❌ WRONG - Test next to the code it tests:
internal/app/services/event_service_test.go
```

**Rule:** Unit tests live in `tests/unit/<layer>/` in package `<layer>_test`, driving the code
through its exported surface. Fakes are shared across tests and live in `tests/mocks/`, each
one asserting it satisfies the real interface:

```go
var _ repositories.EventRepository = (*MockEventRepository)(nil)
```

Tests that need a real database go in `tests/integration/`.

### 4. Dependency Injection (MANDATORY)

```go
❌ WRONG - Standalone func registered directly:
api.GET("/events", listEvents)   // free function reaching for a global

✅ CORRECT - Constructor-based DI, assembled once in routers/index.go:
// internal/app/routers/index.go
eventRepo := repositories.NewEventRepository()
ticketRepo := repositories.NewTicketRepository()
eventService := services.NewEventService(eventRepo, ticketRepo)

RegisterEventRoutes(apiV1, eventService)

// internal/app/routers/event_routes.go
func RegisterEventRoutes(group *gin.RouterGroup, eventService *services.EventService) {
    eventController := controllers.NewEventController(eventService)
    group.GET("/events", eventController.List)
    group.GET("/events/:id", eventController.Get)
}
```

**Rule:** Use `New*` constructors for dependency injection. `internal/app/routers/index.go` is
the **single** place that constructs concrete repositories and services; every feature then gets
its own `Register<Name>Routes` file.

---

## 🔥 TIER 1: HARD LIMITS (EXCEED = REJECT CODE)

```
File Size:     MAX 300 lines  (warning at 250)
Function Size: MAX 100 lines  (warning at 80)
Test Coverage: MIN 70% for services
```

---

## 📍 TIER 2: CRITICAL PATTERNS

### Response Utilities (pkg/utils/response.go)

```go
// Success responses
utils.Ok(c, data, message)              // 200
utils.Created(c, data, message)          // 201
utils.NoContent(c)                       // 204

// Error responses
utils.BadRequest(c, err, message)        // 400
utils.Unauthorized(c, err, message)      // 401
utils.Forbidden(c, err, message)         // 403
utils.NotFound(c, err, message)          // 404
utils.Conflict(c, err, message)          // 409
utils.InternalServerError(c, err, msg)   // 500
```

### Error Handling

```go
❌ WRONG:
_, _ = someFunc()  // Ignored error
if err != nil {
    return
}

✅ CORRECT:
result, err := someFunc()
if err != nil {
    logger.Errorf("operation failed: %v", err)
    return fmt.Errorf("failed to do X: %w", err)
}
```

Services declare sentinel errors; controllers translate them into responses:

```go
// internal/app/services/event_service.go
var ErrEventNotFound = errors.New("event not found")

// internal/app/controllers/event_controller.go
if errors.Is(err, services.ErrEventNotFound) {
    utils.NotFound(c, err, "Event not found")
    return
}
```

### Logging

```go
❌ WRONG:
log.Printf("User created")
fmt.Println("Error:", err)

✅ CORRECT:
logger.Infof("user created: ID=%d, Email=%s", user.ID, user.Email)
logger.Errorf("failed to create user: %v", err)
logger.Warnf("approaching rate limit: %d/%d", current, limit)
```

### Routing (one file per feature, one central assembly point)

```go
❌ WRONG - Everything inline in the central router:
func RegisterRoutes(router *gin.Engine) {
    router.GET("/events", func(c *gin.Context) { /* ... */ })
    // ... grows to 500+ lines as features pile up
}

✅ CORRECT - Central router delegates to a per-feature Register function:
// internal/app/routers/index.go
func RegisterRoutes(route *gin.Engine) {
    RegisterHealthRoutes(route)              // root-level probes

    apiV1 := route.Group("/api/v1")
    apiV1.Use(middlewares.RateLimitMiddleware())

    RegisterAuthRoutes(apiV1, authService)
    RegisterEventRoutes(apiV1, eventService)
}
```

**Rules:**
- Each feature owns a `internal/app/routers/<name>_routes.go` with
  `Register<Name>Routes(group *gin.RouterGroup, ...)`.
- `index.go` builds the dependencies and calls those functions — it is the only file that
  changes when a feature is added.
- System probes (`/health`, `/metrics`) mount at the **root** via `RegisterHealthRoutes(route)`,
  not under `/api/v1`.

### Database Access

```go
❌ WRONG - Reaching for the database from a service or controller:
func (s *EventService) List(ctx context.Context) { database.DB.Find(&events) }

✅ CORRECT - Only repositories touch the database:
func (r *eventRepo) List() ([]*models.Event, error) {
    var events []*models.Event
    if err := database.DB.Order("sale_starts_at ASC").Find(&events).Error; err != nil {
        logger.Errorf("failed to list events: %v", err)
        return nil, fmt.Errorf("failed to list events: %w", err)
    }
    return events, nil
}
```

Repositories use the package-level handle in `internal/adapters/database`. A row that does
not exist is `(nil, nil)` — the caller decides whether that is an error.

### Migrations

Schema changes are **versioned SQL**, applied by golang-migrate at startup, and fatal on
failure. There is no AutoMigrate.

```
internal/adapters/database/migrations/sql/
  000005_create_events_table.up.sql
  000005_create_events_table.down.sql
```

Never edit a migration that has already been applied — add a new pair.

### Request Tracing with LogStart/LogFinish (MANDATORY)

```go
func (ctrl *EventController) Get(c *gin.Context) {
    ctx, start := logger.LogStart(c.Request.Context(), "EventController.Get")

    event, err := ctrl.service.Get(ctx, id)
    if err != nil {
        logger.LogFinish(ctx, "EventController.Get", err, start)
        utils.NotFound(c, err, "Event not found")
        return
    }
    logger.LogFinish(ctx, "EventController.Get", nil, start)
    utils.Ok(c, event, "Event retrieved successfully")
}
```

**Span name convention:** `<Type>.<Method>` — e.g. `AuthController.Login`, `EventService.List`.

---

## 📁 File Structure Reference

See [MODULE_GUIDE.md](./MODULE_GUIDE.md) for the full tree. The essentials:

```
.
├── main.go                            → config → db → migrate → seed (dev) → serve → shutdown
├── internal/
│   ├── adapters/database/             → connection, migrations, seeders
│   │   ├── database.go                →   DbConnection(), GetDB(), package-level DB
│   │   ├── migrations/sql/            →   versioned .up.sql / .down.sql pairs
│   │   └── seeders/                   →   idempotent demo data, development only
│   ├── app/
│   │   ├── controllers/               →   HTTP layer: <name>_controller.go
│   │   ├── dto/                       →   request/response types
│   │   ├── middlewares/               →   auth, cors, metrics, rate_limit, request_id, request_log
│   │   ├── routers/                   →   router.go, index.go, <name>_routes.go, swagger.go
│   │   └── services/                  →   business logic: <name>_service.go (auth/ is a package)
│   └── domain/
│       ├── models/                    →   GORM models with TableName()
│       └── repositories/              →   interface + unexported impl + New*Repository()
├── pkg/                               → cross-cutting kit; MUST never import internal/
│   └── config/ logger/ metrics/ types/ utils/   (utils/response.go → MUST use these)
└── tests/
    ├── unit/{controllers,services,middlewares}/  → package <x>_test, no database
    ├── integration/{api,database}/               → needs a real database
    └── mocks/                                    → shared in-memory fakes
```

---

## ⚡ Quick Decision Tree

```
Writing a controller?
  → Struct + injected service + response utils? YES → ✅
  → Standalone func / business logic in the controller? ❌ STOP

Writing a service?
  → Struct + repository INTERFACE injected via New*? YES
  → Unit test in tests/unit/services with a fake from tests/mocks? YES → ✅
  → Touching database.DB directly? ❌ STOP

Writing a repository?
  → Exported interface + unexported struct + New*Repository()? YES
  → Errors wrapped with %w and logged? YES → ✅
  → Returning gorm.ErrRecordNotFound to the caller? ❌ STOP — return (nil, nil)

Returning a response?
  → Using utils.Ok/Created/NotFound/...? YES → ✅
  → Using c.JSON directly? ❌ STOP

Adding routes?
  → New <name>_routes.go + one call from index.go? YES → ✅
  → Handlers written inline in index.go? ❌ STOP

Changing the schema?
  → New versioned .up.sql + .down.sql pair? YES → ✅
  → Editing an applied migration, or reaching for AutoMigrate? ❌ STOP

File approaching 250 lines?
  → Split into another file in the SAME package? YES → ✅
  → Keep adding? ❌ STOP
```

---

## Git: Do not commit non-essential .md files

**Rule:** Markdown files that are **local, analysis-only, or temporary** must not be committed or pushed.

**Do not commit (examples):**
- `PROJECT_ANALYSIS.md` — project analysis/score (local only)
- Draft docs, personal notes, or .md used only for internal reference and not part of the shared project

**Do commit:** All files in `docs/` that are part of the project standard (CODING_STANDARDS, DESIGN_PATTERNS, OBSERVABILITY, CONFIGURATION, AI rules, README, .env.example, etc.).

Non-essential files are listed in `.gitignore` (e.g. `PROJECT_ANALYSIS.md`). Before committing, ensure no new analysis/local .md files are staged.

---

## 📚 For More Details

- Layout source of truth: [`MODULE_GUIDE.md`](./MODULE_GUIDE.md)
- Full standards: [`CODING_STANDARDS.md`](./CODING_STANDARDS.md) (read the sections marked CRITICAL)
- Design patterns: [`DESIGN_PATTERNS.md`](./DESIGN_PATTERNS.md)
- Quick templates: [`AI_QUICK_REFERENCE.md`](./AI_QUICK_REFERENCE.md)

**When in doubt, copy the `event` slice** — it is the most recent feature and follows every
rule on this page:

```
internal/adapters/database/migrations/sql/000005_create_events_table.up.sql
internal/domain/models/event_model.go
internal/domain/repositories/event_repo.go
internal/app/dto/event_dto.go
internal/app/services/event_service.go
internal/app/controllers/event_controller.go
internal/app/routers/event_routes.go
tests/mocks/event_repo_mock.go
tests/unit/services/event_service_test.go
```

---

**Remember:** These are COMPANY STANDARDS. Violation = Code Rejected.
