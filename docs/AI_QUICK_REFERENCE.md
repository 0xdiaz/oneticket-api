# AI Agent Quick Reference

> **Print this mentally before every code change!**

> 🧭 **Structure:** Code is organized by technical layer —
> `internal/app/controllers` → `internal/app/services` → `internal/domain/repositories` →
> `internal/domain/models`, wired in `internal/app/routers/index.go`. See
> **[MODULE_GUIDE.md](./MODULE_GUIDE.md)** (the source of truth for the layout). The templates
> below are the real shapes from the `event` slice — copy that slice to start a new feature.

---

## ⚠️ FIRST TIME HERE?

**🚨 READ [`00_AI_CRITICAL_RULES.md`](./00_AI_CRITICAL_RULES.md) FIRST!**

That file contains the absolute non-negotiable rules.
This file is for quick templates and checklists.

---

## ⚡ THE 5 COMMANDMENTS

```
1. 📏 File >300 lines?        → STOP. Split it.
2. 📐 Function >100 lines?    → STOP. Extract functions.
3. 🧪 No tests?               → STOP. Write tests first.
4. ❌ Error ignored (_, _)?   → STOP. Handle it.
5. 📝 Exported without docs?  → STOP. Document it.
```

**VIOLATE = CODE REJECTED**

---

## 🎯 Before Writing ANY Code

```bash
# Ask yourself:
□ Which layer am I in? (controller / service / repository / model)
□ Am I following dependency direction? (controller → service → repository → model)
□ Does my service depend on a repository INTERFACE, not a concrete type?
□ Will this file exceed 300 lines? → Plan to split (another file, SAME package)
□ Will this function exceed 100 lines? → Plan to extract
□ Do I need tests? → Yes, ALWAYS for services (tests/unit/services + a fake in tests/mocks)
□ Is this documented? → Required for exported items
```

---

## 📐 Size Limits (HARD LIMITS)

```
File:     MAX 300 lines  (warning at 250)
Function: MAX 100 lines  (warning at 80)
```

**Approaching limit?**
- Stop and refactor NOW
- Don't wait until you exceed
- Split proactively

---

## 🏗️ Architecture Cheat Sheet

```
Request flow:
routers/index.go → Register<Name>Routes → Controller → Service → Repository → Database

Layers (one file per feature, per directory):
┌─────────────────────────────────────────────────────────────────┐
│ Controller  │ internal/app/controllers/<name>_controller.go     │  HTTP only
│ Service     │ internal/app/services/<name>_service.go           │  business logic
│ Repository  │ internal/domain/repositories/<name>_repo.go       │  data access only
│ Model       │ internal/domain/models/<name>_model.go            │  GORM struct
│ DTO         │ internal/app/dto/<name>_dto.go                    │  request/response
│ Routes      │ internal/app/routers/<name>_routes.go             │  route registration
└─────────────────────────────────────────────────────────────────┘

Dependency direction (never upward):
Controller  →  Service  →  Repository  →  Model
```

**Never:**
- ❌ Controller with business logic
- ❌ Controller calling a repository directly
- ❌ Service touching `database.DB`
- ❌ Repository importing a service

**Reference slice:** `event` — copy it to start a new feature. Use the repo's actual import
path (`github.com/0xdiaz/oneticket-api/...`).

---

## 🔥 Layer Templates

### `<name>_model.go` — the GORM struct

```go
// internal/domain/models/event_model.go
package models

import "time"

// Event is a sellable event.
type Event struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name" gorm:"type:varchar(200);not null"`

    // PriceCents is the ticket price in the smallest currency unit.
    // Money is an integer on purpose: floats lose cents under arithmetic.
    PriceCents int64 `json:"price_cents" gorm:"not null"`

    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the database table name for Event model.
func (e *Event) TableName() string {
    return "events"
}
```

Every model needs a matching **versioned migration** —
`internal/adapters/database/migrations/sql/NNNNNN_create_<table>.up.sql` plus its `.down.sql`.
There is no AutoMigrate.

### `<name>_dto.go` — request/response types

```go
// internal/app/dto/event_dto.go
package dto

// EventResponse is the API representation of an event.
type EventResponse struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`

    // AvailableTickets is counted live from the tickets table.
    AvailableTickets int64 `json:"available_tickets"`
}
```

Request DTOs carry `binding:"..."` tags and are bound at the HTTP boundary with
`c.ShouldBindJSON(&req)`.

### `<name>_repo.go` — data access, the ONLY layer touching the database

```go
// internal/domain/repositories/event_repo.go
package repositories

// EventRepository defines data access for events.
type EventRepository interface {
    List() ([]*models.Event, error)
    // GetByID returns the event, or (nil, nil) when it does not exist.
    GetByID(id uint) (*models.Event, error)
}

type eventRepo struct{}

// NewEventRepository returns a new EventRepository implementation.
func NewEventRepository() EventRepository {
    return &eventRepo{}
}

func (r *eventRepo) GetByID(id uint) (*models.Event, error) {
    var event models.Event
    err := database.DB.Where("id = ?", id).First(&event).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // a missing row is not an error at this layer
        }
        logger.Errorf("failed to get event by id: %v", err)
        return nil, fmt.Errorf("failed to get event by id: %w", err)
    }
    return &event, nil
}
```

### `<name>_service.go` — business logic + sentinel errors

```go
// internal/app/services/event_service.go
package services

// ErrEventNotFound is returned when the requested event does not exist.
var ErrEventNotFound = errors.New("event not found")

// EventService handles event and ticket availability business logic.
type EventService struct {
    eventRepo  repositories.EventRepository
    ticketRepo repositories.TicketRepository
}

// NewEventService creates a new EventService instance.
func NewEventService(eventRepo repositories.EventRepository, ticketRepo repositories.TicketRepository) *EventService {
    return &EventService{eventRepo: eventRepo, ticketRepo: ticketRepo}
}

// Get returns a single event with its live ticket availability.
//
// Returns ErrEventNotFound when no event has the given id.
func (s *EventService) Get(ctx context.Context, id uint) (response *dto.EventResponse, err error) {
    ctx, start := logger.LogStart(ctx, "EventService.Get")

    event, err := s.eventRepo.GetByID(id)
    if err != nil {
        logger.LogFinish(ctx, "EventService.Get", err, start)
        return nil, fmt.Errorf("failed to get event: %w", err)
    }
    if event == nil {
        logger.LogFinish(ctx, "EventService.Get", ErrEventNotFound, start)
        return nil, ErrEventNotFound
    }

    logger.LogFinish(ctx, "EventService.Get", nil, start)
    return buildResponse(event), nil
}
```

Services take repository **interfaces** — that is what makes them testable without a database.

### `<name>_controller.go` — thin HTTP layer

```go
// internal/app/controllers/event_controller.go
package controllers

// EventController handles event browsing endpoints.
type EventController struct {
    service *services.EventService
}

// NewEventController creates a new EventController instance.
func NewEventController(service *services.EventService) *EventController {
    return &EventController{service: service}
}

// Get returns a single event with live ticket availability.
//
// GET /api/v1/events/:id
func (ctrl *EventController) Get(c *gin.Context) {
    ctx, start := logger.LogStart(c.Request.Context(), "EventController.Get")

    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        logger.LogFinish(ctx, "EventController.Get", err, start)
        utils.BadRequest(c, err, "Invalid event id")
        return
    }

    event, err := ctrl.service.Get(ctx, uint(id))
    if err != nil {
        if errors.Is(err, services.ErrEventNotFound) {
            logger.LogFinish(ctx, "EventController.Get", err, start)
            utils.NotFound(c, err, "Event not found")
            return
        }
        logger.Errorf("failed to get event: %v", err)
        logger.LogFinish(ctx, "EventController.Get", err, start)
        utils.InternalServerError(c, err, "Failed to get event")
        return
    }

    logger.LogFinish(ctx, "EventController.Get", nil, start)
    utils.Ok(c, event, "Event retrieved successfully")
}
```

### `<name>_routes.go` — route registration for one feature

```go
// internal/app/routers/event_routes.go
package routers

// RegisterEventRoutes registers event browsing routes under the given group.
func RegisterEventRoutes(group *gin.RouterGroup, eventService *services.EventService) {
    eventController := controllers.NewEventController(eventService)
    group.GET("/events", eventController.List)
    group.GET("/events/:id", eventController.Get)
}
```

### Wire it — a few lines in `routers/index.go`

```go
// internal/app/routers/index.go — the only file that knows concrete types
func RegisterRoutes(route *gin.Engine) {
    RegisterHealthRoutes(route) // /health and /metrics at the ROOT

    apiV1 := route.Group("/api/v1")
    apiV1.Use(middlewares.RateLimitMiddleware())

    eventRepo := repositories.NewEventRepository()
    ticketRepo := repositories.NewTicketRepository()
    eventService := services.NewEventService(eventRepo, ticketRepo)

    RegisterEventRoutes(apiV1, eventService)

    // Protected routes sit behind the auth middleware.
    protected := apiV1.Group("")
    protected.Use(middlewares.AuthMiddleware(authService))
}
```

**Rules:**
- ✅ Each layer's `New*` takes its dependency; `index.go` assembles the chain.
- ✅ Adding a feature touches `index.go` and adds one `<name>_routes.go`.
- ❌ Never register handlers inline in `index.go`.

---

## ✅ Error Handling Pattern

```go
// ✅ ALWAYS do this:
result, err := someFunction()
if err != nil {
    logger.Errorf("context: %v", err)                    // Log
    return fmt.Errorf("operation failed: %w", err)       // Wrap with %w
}

// ❌ NEVER do this:
result, _ := someFunction()                              // Ignored!
result, err := someFunction()
if err != nil {
    panic(err)                                           // Panic!
}
result, err := someFunction()
return err                                               // Not wrapped!
```

**Sentinel errors** are declared in the service and translated in the controller:

```go
// service
var ErrEventNotFound = errors.New("event not found")

// controller
if errors.Is(err, services.ErrEventNotFound) {
    utils.NotFound(c, err, "Event not found")
    return
}
```

The `auth` service package keeps its own sentinels (`auth.ErrUserNotFound`,
`auth.ErrInvalidCredentials`, …) and maps them centrally in `authErrToAPIError`.

---

## 📝 Documentation Pattern

```go
// ✅ CORRECT:
// Get returns a single event with its live ticket availability.
//
// Returns ErrEventNotFound when no event has the given id.
func (s *EventService) Get(ctx context.Context, id uint) (*dto.EventResponse, error) {
    // implementation
}

// ❌ WRONG:
// Get event
func (s *EventService) Get(ctx context.Context, id uint) (*dto.EventResponse, error) {

// ❌ WRONG:
func (s *EventService) Get(ctx context.Context, id uint) (*dto.EventResponse, error) {  // No comment
```

---

## 🧪 Testing Checklist

```
⚠️  Unit tests live in the tests/ tree, NOT next to the code
□ Create tests/unit/services/<name>_service_test.go
□ Use package services_test (black box) — drive the exported surface
□ Create a shared fake in tests/mocks/<name>_repo_mock.go
□ Assert the fake satisfies the interface:
     var _ repositories.EventRepository = (*MockEventRepository)(nil)
□ Test happy path
□ Test 2+ error cases (drive the fake's error fields)
□ Use table-driven subtests if >3 scenarios
□ Assert with testify (require for fatal, assert for the rest)
□ Run: make test   (or: go test ./tests/unit/...)
```

**Canonical fake + test** (from `tests/mocks/event_repo_mock.go` and
`tests/unit/services/event_service_test.go`):

```go
// tests/mocks/event_repo_mock.go
package mocks

// MockEventRepository is an in-memory EventRepository for unit tests.
type MockEventRepository struct {
    mu     sync.RWMutex
    nextID uint
    byID   map[uint]*models.Event

    // GetErr, when set, is returned instead of data.
    GetErr error
}

var _ repositories.EventRepository = (*MockEventRepository)(nil)
```

```go
// tests/unit/services/event_service_test.go
package services_test

func TestEventServiceGet(t *testing.T) {
    eventRepo := mocks.NewMockEventRepository()
    ticketRepo := mocks.NewMockTicketRepository()
    event := eventRepo.Seed(&models.Event{Name: "Flash Sale Demo", TotalTickets: 100})
    ticketRepo.SeedAvailable(event.ID, 100)

    service := services.NewEventService(eventRepo, ticketRepo)

    got, err := service.Get(context.Background(), event.ID)

    require.NoError(t, err)
    assert.Equal(t, int64(100), got.AvailableTickets)
}
```

Tests that need a real database go in `tests/integration/`.

---

## 🚨 Forbidden Patterns

```go
❌ panic() in business logic
❌ _, _ = someFunc()                    // Ignored error
❌ "SELECT * FROM " + table             // SQL injection
❌ if x { if y { if z { } } }           // Too nested (>3 levels)
❌ password := "hardcoded"              // Hardcoded secrets
❌ log.Printf()                         // Use logger.Infof()
❌ file size >300 lines
❌ function >100 lines
❌ c.JSON(...) for API responses        // Use pkg/utils (utils.Ok, utils.BadRequest, …)
❌ database.DB in a service/controller  // Only repositories touch the database
❌ returning gorm.ErrRecordNotFound     // Return (nil, nil); let the service decide
❌ editing an already-applied migration // Add a new .up.sql / .down.sql pair
❌ float types for money                // Use int64 in the smallest currency unit
❌ No tests for services
❌ Exported function without docs
```

---

## 🎨 Naming Conventions

```go
// Files (feature-prefixed, snake_case)
✅ event_service.go, event_repo.go, auth_service_tokens.go
❌ EventService.go, event-service.go, eventService.go

// Packages (named after the layer directory)
✅ package controllers, package services, package repositories, package models
✅ package auth  (a service that needs multiple files gets its own subpackage)
❌ package Services, package event_svc

// Types
✅ EventController, EventService, EventRepository, MockEventRepository
❌ EventCtrl, EventSvc, EventRepoImpl

// Constructors
✅ NewEventService, NewEventRepository, NewEventController
❌ CreateEventService, MakeEventRepo

// Variables
✅ event, eventID, ticketRepo
❌ e, evtID, ticket_repo

// Constants
✅ const TicketStatusAvailable = "available"
❌ const TICKET_STATUS_AVAILABLE = "available"
```

---

## 🔍 Pre-Commit Checklist

```bash
□ All functions <100 lines?
□ All files <300 lines?
□ All errors handled and wrapped with %w?
□ All exported items documented?
□ Tests written and passing?
□ No hardcoded secrets?
□ No panic() in business logic?
□ No SQL string concatenation?
□ No ignored errors (_, _)?
□ New migration has a matching .down.sql?
□ gofmt applied?

# Run these:
gofmt -w .
go vet ./...
go build ./...
make test
```

---

## 💡 Common Patterns

### Transactions

Multi-step writes that must succeed or fail together go through a single GORM transaction,
inside the repository layer:

```go
func (r *eventRepo) DoTwoThings(...) error {
    return database.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&a).Error; err != nil {
            return fmt.Errorf("create a: %w", err)
        }
        if err := tx.Model(&b).Update("x", y).Error; err != nil {
            return fmt.Errorf("update b: %w", err)
        }
        return nil
    })
}
```

Returning a non-nil error from the callback rolls the whole thing back.

### Validation (gin binding tags, bound at the HTTP boundary)

```go
// internal/app/dto/event_dto.go
type CreateEventRequest struct {
    Name       string `json:"name" binding:"required,max=200"`
    PriceCents int64  `json:"price_cents" binding:"required,min=0"`
}

// internal/app/controllers/event_controller.go
var req dto.CreateEventRequest
if err := c.ShouldBindJSON(&req); err != nil {
    utils.BadRequest(c, err, "Invalid request data")
    return
}
```

### DataTables (server-side pagination/search/sort)

See `internal/domain/repositories/example_repo.go` for the wiring of the
`Datatables-Gin` helper.

---

## 📚 Quick Links

- [`00_AI_CRITICAL_RULES.md`](./00_AI_CRITICAL_RULES.md) — the non-negotiables, read first
- [`MODULE_GUIDE.md`](./MODULE_GUIDE.md) — layout source of truth
- [`CODING_STANDARDS.md`](./CODING_STANDARDS.md) — full standards
- [`DESIGN_PATTERNS.md`](./DESIGN_PATTERNS.md) — patterns and rationale
- [`AUTHENTICATION.md`](./AUTHENTICATION.md) — JWT, refresh rotation, password reset
- [`MIGRATIONS.md`](./MIGRATIONS.md) — schema change workflow

---

## 🎯 Remember

**When in doubt, copy the `event` slice.** It is the most recent feature and follows every
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
