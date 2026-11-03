# AGENT.md

## 1. Purpose

You are a software development agent responsible for designing, implementing, testing, refactoring, and maintaining this Go project.

All code changes MUST follow these principles:

* **Test-Driven Development (TDD)**
* **SOLID principles**
* **Hexagonal Architecture (Ports and Adapters)**
* Clear separation of responsibilities
* Low coupling
* High cohesion
* Maintainability
* Testability
* Modern Go idioms and best practices
* Explicit and controlled dependencies

Code quality and architectural integrity take priority over implementation speed.

---

## 2. Core Architectural Principles

The project follows **Hexagonal Architecture**, also known as **Ports and Adapters Architecture**.

The architecture must protect the application core from external technical concerns.

The core application must not depend directly on:

* Web frameworks (e.g., Gin, Fiber, Echo, Chi)
* Databases or ORMs (e.g., GORM, sqlx, `database/sql` drivers directly in domain)
* External HTTP APIs
* File systems
* Message brokers (e.g., RabbitMQ, Kafka, NATS)
* CLI frameworks (e.g., Cobra, Urfave/cli)
* Specific infrastructure libraries

External technologies must be connected through well-defined ports (interfaces) and adapters.

---

## 3. Recommended Project Structure

Use a structure compliant with standard Go project layouts:

```text
internal/
├── domain/
│   ├── entities/
│   ├── value_objects/
│   ├── services/
│   ├── errors/
│   └── repositories/
│
├── application/
│   ├── commands/
│   ├── queries/
│   ├── usecases/
│   └── dto/
│
├── ports/
│   ├── inbound/
│   └── outbound/
│
└── adapters/
    ├── inbound/
    │   ├── http/
    │   ├── cli/
    │   └── ...
    │
    └── outbound/
        ├── persistence/
        ├── external_services/
        └── ...

```

Tests should follow the same architectural boundaries:

```text
tests/
├── unit/
│   ├── domain/
│   └── application/
│
├── integration/
│   └── adapters/
│
└── e2e/

```

The exact directory structure may be adapted to the project, but the architectural boundaries MUST remain explicit.

### Current GRZ package mapping

The current implementation keeps those boundaries in these packages:

```text
internal/domain/kb                          # KB entity and domain rules
internal/application/usecases                # KB use-case orchestration
internal/application/dto                     # application input/output models
internal/application/ports/{inbound,outbound} # application contracts
internal/adapters/outbound/persistence       # Markdown file repository
web                                          # inbound HTTP adapter
html/pages                                   # HTML presentation adapter
main.go                                      # dependency composition root
```

Persistence formats and filesystem access belong in adapters; domain packages
must not serialize entities or calculate presentation-specific layout. HTTP
handlers depend on the inbound application port, never directly on a repository.

---

## 4. Dependency Rule

Dependencies MUST always point toward the application core.

```text
Adapters
    ↓
Ports
    ↓
Application
    ↓
Domain

```

The domain must remain independent of infrastructure.

### Forbidden

```go
// internal/domain/user.go

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

```

### Preferred

Go interfaces define the ports directly in the application or domain layers:

```go
// internal/application/ports/outbound/user_repository.go
package outbound

import (
    "context"
    "yourproject/internal/domain/entities"
)

type UserRepository interface {
    FindByID(ctx context.Context, id entities.UserID) (*entities.User, error)
    Save(ctx context.Context, user *entities.User) error
}

```

The concrete implementation belongs to an adapter:

```go
// internal/adapters/outbound/persistence/postgres_user_repository.go
package persistence

import (
    "context"
    "database/sql"
    "yourproject/internal/domain/entities"
)

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *entities.User) error {
    // Concrete SQL implementation
    return nil
}

```

---

## 5. Domain Layer

The domain contains the core business rules.

It should contain concepts such as:

* Entities (structs with business logic methods)
* Value Objects (immutable structs)
* Domain Services
* Domain Errors
* Domain Events
* Domain-specific interfaces

The domain MUST NOT contain infrastructure concerns.

The domain should be:

* Framework-independent
* Database-independent
* Highly testable
* Deterministic whenever possible

Example:

```go
package valueobjects

import "errors"

var ErrCurrencyMismatch = errors.New("currencies do not match")
var ErrNegativeAmount = errors.New("amount cannot be negative")

type Money struct {
    amount   int64 // stored in smallest currency unit (e.g. cents)
    currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    return Money{amount: amount, currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

```

---

## 6. Application Layer

The application layer orchestrates use cases.

A use case should:

1. Receive a command or input DTO.
2. Retrieve required data through ports.
3. Invoke domain behavior.
4. Persist changes through ports.
5. Return an appropriate result or error.

Example:

```go
package usecases

import (
    "context"
    "yourproject/internal/application/dto"
    "yourproject/internal/application/ports/outbound"
    "yourproject/internal/domain/entities"
)

type CreateUserUseCase struct {
    repo outbound.UserRepository
}

func NewCreateUserUseCase(repo outbound.UserRepository) *CreateUserUseCase {
    return &CreateUserUseCase{repo: repo}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, cmd dto.CreateUserCommand) (*entities.User, error) {
    user, err := entities.NewUser(cmd.Name, cmd.Email)
    if err != nil {
        return nil, err
    }

    if err := uc.repo.Save(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

```

A use case MUST NOT:

* Create database connections directly.
* Call external HTTP APIs directly.
* Depend directly on an ORM or SQL driver.
* Depend directly on a web framework or HTTP request/response objects.
* Contain infrastructure-specific logic.

---

## 7. Ports

Ports define contracts between the application core and external systems using Go interfaces.

### 7.1 Inbound Ports

Inbound ports define what the application exposes.

Example:

```go
type CreateUserUseCase interface {
    Execute(ctx context.Context, cmd dto.CreateUserCommand) (*entities.User, error)
}

```

Inbound adapters call inbound ports.

```text
HTTP Controller / REST API
CLI Command
Message Consumer
Scheduled Job (Cron)
       ↓
Inbound Port
       ↓
Use Case

```

---

### 7.2 Outbound Ports

Outbound ports define what the application requires from external systems.

Example:

```go
type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByID(ctx context.Context, id entities.UserID) (*entities.User, error)
}

```

The application defines the contract. The infrastructure implements it.

```text
Application Core
       ↓
UserRepository (Interface)
       ↑
PostgresUserRepository (Adapter Struct)

```

---

## 8. Adapters

Adapters connect the application to the outside world.

### 8.1 Inbound Adapters

Examples:

* HTTP / REST Handlers
* gRPC Controllers
* CLI Handlers
* Message consumers (AMQP/Kafka listeners)
* Scheduled jobs

Their responsibility is to translate external input into application commands/queries, invoke the use case, and serialize output.

Controllers MUST NOT contain business logic.

---

### 8.2 Outbound Adapters

Examples:

* PostgreSQL / SQLite implementations
* Redis cache adapters
* External HTTP API clients
* File system storage
* Event publishers

They implement outbound port interfaces.

```text
Use Case
   ↓
Repository Port Interface
   ↓
Persistence Adapter
   ↓
Database (sql.DB)

```

---

## 9. Test-Driven Development

TDD is mandatory for new behavior.

Follow the cycle:

```text
RED
 ↓
GREEN
 ↓
REFACTOR

```

### RED

Write a test describing the expected behavior before implementing the functionality. The test must fail for the expected reason.

Example:

```go
func TestUserCannotBeCreatedWithInvalidEmail(t *testing.T) {
    _, err := entities.NewUser("John", "invalid-email")
    if !errors.Is(err, entities.ErrInvalidEmail) {
        t.Errorf("expected ErrInvalidEmail, got %v", err)
    }
}

```

### GREEN

Implement the minimum amount of code required to make the test pass. Do not implement speculative functionality.

### REFACTOR

Once the tests are green:

* Remove duplication.
* Improve naming.
* Simplify the implementation.
* Apply SOLID principles.
* Verify architectural boundaries.

Tests must remain green after refactoring.

---

## 10. Testing Strategy

Organize tests using standard Go tools (`go test`).

Preferred test pyramid:

```text
       E2E
      /   \
  Integration
  /         \
  Unit Tests

```

### 10.1 Unit Tests

Unit tests should cover:

* Domain entities & value objects
* Domain services
* Use cases

External dependencies should be replaced with simple **fakes** or mock implementations. Prefer hand-crafted mock structs or `testify/mock` over heavy mocking frameworks.

Use Go table-driven tests (`struct{ name string; ... }`) for test case coverage.

For performance-sensitive code, add benchmarks and compare before/after using
`go test <package> -run '^$' -bench . -benchmem`. Keep optimizations that improve
representative workloads without changing correctness or observable ordering.

---

### 10.2 Integration Tests

Integration tests verify that adapters interact correctly with external systems.

Examples:

```text
Repository Adapter + Real DB / Testcontainers
HTTP Handler + Application Layer (httptest.NewServer)
External Service Adapter + Mock HTTP Server

```

Do not mock away the exact dependency that an integration test aims to verify.

---

### 10.3 End-to-End Tests

E2E tests verify full user journeys across the entire system. Keep E2E tests limited to critical flows.

---

## 11. SOLID Principles

All design decisions should respect SOLID:

* **S — Single Responsibility Principle:** A struct or package should have one clear responsibility. Avoid monolithic handler or service structs.
* **O — Open/Closed Principle:** Extend capabilities by injecting interface implementations rather than modifying existing code blocks.
* **L — Liskov Substitution Principle:** Any struct implementing a port interface must be substitutable without breaking expectations or contract semantics.
* **I — Interface Segregation Principle:** Keep interfaces small and client-focused (e.g., single-method interfaces like `Reader`, `Writer` or specific domain interfaces).
* **D — Dependency Inversion Principle:** High-level application code depends on abstractions (Go interfaces), not on concrete structs or external packages.

---

## 12. Dependency Injection

Dependencies MUST be explicitly injected (usually via constructor functions `New...`), never instantiated directly inside business logic.

Preferred:

```go
type OrderService struct {
    repo    OrderRepository
    payment PaymentGateway
}

func NewOrderService(repo OrderRepository, payment PaymentGateway) *OrderService {
    return &OrderService{
        repo:    repo,
        payment: payment,
    }
}

```

Avoid:

```go
func NewOrderService() *OrderService {
    return &OrderService{
        repo:    persistence.NewPostgresOrderRepo(), // Direct concrete binding
    }
}

```

Concrete dependency resolution and wiring MUST happen in a **Composition Root** (typically in `cmd/main.go`).

---

## 13. Framework Isolation

Frameworks and infrastructure libraries are implementation details.

The core packages (`internal/domain`, `internal/application`) MUST NOT import:

```text
net/http (except context definitions if needed, preferred context package only)
github.com/gin-gonic/gin
github.com/labstack/echo
gorm.io/gorm
github.com/jmoiron/sqlx
github.com/go-redis/redis

```

The goal is to allow framework or infrastructure replacement without altering core domain logic.

---

## 14. Idiomatic Go & Type Safety

* **Strong Typing:** Use custom types (e.g., `type UserID string`) to prevent primitive obsession.
* **Explicit Error Handling:** Check errors explicitly (`if err != nil`). Never ignore errors using `_`.
* **Context Handling:** Pass `ctx context.Context` as the first parameter in use cases, ports, and IO-bound methods.
* **Avoid Empty Interfaces:** Avoid `interface{}` / `any` unless strictly necessary (e.g., dynamic JSON payloads or generic utilities).
* **Zero-Values:** Design structs and types to have usable or safe zero-values whenever possible.

---

## 15. Error Handling

Business errors should be sentinel errors or custom error types defined within the domain package.

Example:

```go
var ErrUserAlreadyExists = errors.New("user already exists")

```

Infrastructure errors must not leak into the core. Wrap technical errors at the adapter boundary using `fmt.Errorf("adapter error: %w", err)`.

---

## 16. Configuration and Secrets

Configuration must remain separate from code.

Never hard-code:

* API keys
* Database credentials
* Tokens
* Environment-specific URLs

Use environment variables or explicit configuration structs loaded at startup (e.g., via `envconfig` or `viper` in `main.go`).

---

## 17. Code Quality

Code must strictly adhere to standard Go formatting and linting conventions:

* Format code using `gofmt` or `goimports`.
* Run static analysis with `go vet` and `golangci-lint`.
* Write clear, concise standard Go doc comments for exported types and functions.
* Keep functions small and cyclomatic complexity low.

---

## 18. Avoid Premature Abstraction

Do not create an interface for every single struct unless it serves as an explicit port, has multiple implementations, or is required for testing.

Prefer:

```text
Simple + Testable + Decoupled

```

over over-engineered hierarchies.

---

## 19. Development Workflow

For every new feature or change, follow this process:

### Step 1 — Understand

Identify requirements, domain boundaries, existing tests, and ports involved.

### Step 2 — Identify Responsibilities

Assign responsibilities to Domain, Application, Port, or Adapter.

### Step 3 — Write Tests (TDD)

Write unit/integration tests describing expected behavior first.

### Step 4 — Implement

Implement minimal code to pass tests.

### Step 5 — Run Tests & Linters

```bash
go test -v ./...
golangci-lint run

```

### Step 6 — Refactor

Refactor for clarity, idiomatic Go, and SOLID rules while maintaining green tests.

### Step 7 — Add Integration Tests

Add adapter-level integration tests if external systems were touched.

### Step 8 — Update Documentation

Update documentation and `AGENT.md` if conventions or architecture were modified.

### Step 9 — Final Validation

Ensure all items on the checklist pass.

---

## 20. Modifying Existing Code

1. Read existing implementation and tests.
2. Identify architectural boundaries.
3. Add/modify tests first.
4. Implement change and verify tests pass.
5. Refactor without altering public contracts unannounced.
6. Keep documentation up to date.

---

## 21. Documentation Synchronization

Documentation must reflect code implementation.

Verify before completing a task:

```text
[ ] Code matches documentation
[ ] Architecture documentation reflects actual code layout
[ ] Commands documented in README/AGENT.md work
[ ] AGENT.md is updated with any project rule changes

```

---

## 22. AGENT.md Maintenance

`AGENT.md` is a living document. It MUST be updated whenever project conventions, folder structure, testing practices, or rules evolve.

---

## 23. Forbidden Practices

Do not:

* Put business logic in HTTP handlers or database adapters.
* Access database drivers directly inside use cases.
* Import web or database packages inside `internal/domain`.
* Ignore returned errors (`_ = operation()`).
* Use global mutable variables or unchecked singletons.
* Skip TDD or unit tests.
* Leave `AGENT.md` or architectural documentation outdated.

---

## 24. Definition of Done

A task is complete only when:

```text
[ ] Expected behavior is covered by tests
[ ] TDD workflow was followed
[ ] `go test ./...` passes without errors
[ ] `golangci-lint run` passes cleanly
[ ] Hexagonal architecture boundaries are respected
[ ] Core layers remain framework- and DB-independent
[ ] Dependencies are explicitly injected
[ ] Code follows idiomatic Go standards
[ ] No secrets or credentials are hard-coded
[ ] Project documentation & AGENT.md are fully up to date

```

---

## 25. Final Decision Rules

Prioritize:

1. **Correctness**
2. **Business Rules Integrity**
3. **Testability**
4. **Simplicity & Idiomatic Go**
5. **Decoupling**

---

## 26. Server Instances & Port 8080 (Development Hygiene)

The app binds permanently to `:8080` (`http.ListenAndServe(":8080", mux)` in `main.go`).

Rules for any agent session or human developer testing the server locally:

1. **Before starting** the server, free the port if an old instance holds it:
   `fuser -k 8080/tcp` (or `pkill -f "grz"` for a previous build), and check with
   `ss -tln | grep 8080`.
2. **Never leave an instance running** at the end of a session. Explicitly kill
   it: `fuser -k 8080/tcp; pkill -f "/path/to/grz"` — an orphaned process keeps
   port 8080 occupied and silently breaks the next run (the startup log only
   shows "port 8080 already in use" failure at exit).
3. Prefer building a test binary in a scratch directory (e.g. `/tmp`) and
   running `./grz` there, so the project's `data/kb/` directory stays clean:
   test KBs must not pollute the real storage.
4. **After smoke tests**, verify the port is actually released
   (`ss -tln | grep 8080` returns nothing) before declaring the task complete.

---

## 27. Local CSS Generation (Tailwind, No CDN)

The styles come from a **locally generated** stylesheet, embedded in the
binary via `go:embed` (`web/assets.go`, served at `GET /assets/tailwind.css`).
There is no Tailwind CDN or Node toolchain.

- Source of truth for colors: `html/pages/palette.go` (Solarized constants).
  `tailwind.config.js` mirrors them — tests (`html/pages/palette_test.go`)
  fail if the two drift apart.
- Utility classes live inside Go source strings (`html/**`, `web/**`,
  `main.go`), scanned by the content globs in `tailwind.config.js`.
- **Any change adding or renaming a utility class in Go requires
  regenerating the stylesheet:**

  ```bash
  ./tools/build-css.sh          # one-shot generation (CLI standalone, pinned)
  ./tools/build-css.sh --watch  # watch mode during development
  ```

  The script caches the `tailwindcss` standalone CLI (no Node) in
  `$HOME/.cache/grz` and writes the minified output to
  `web/static/tailwind.css`, which is committed with the binary-side embed.
- Page-level styles (body, `.md` markdown sections, SVG graph) remain
  **inline** in `layout.go` / `kb_graph.go`, consuming `var(--g-*)` —
  do not hard-code hex colors in CSS strings; change the values in
  `palette.go` instead.

### Theming & Palette (Phases 3–4)

- Colors are **CSS variables** (`--g-*`), generated in the inline `<style>`
  by `layout.go` (`themeCSSRules`): `:root` = Solarized dark (default),
  `[data-theme="light"]` = Solarized Light ladder (inverted canon).
  `palette.go` holds both maps (`paletteDark`, `paletteLight`) — same token
  set, tests fail on drift (`TestThemePalettes`, `TestLayoutThemeVarRules`).
- Tailwind tokens in `tailwind.config.js` reference `var(--g-*)`; the theme
  is switched at runtime via `<html data-theme="light|dark">` set by the
  **anti-FOUC init script** (place it BEFORE the stylesheet link) and the
  navbar toggle (`#theme-toggle`, persisted in localStorage `grz-theme`).
  Default follows `prefers-color-scheme`.
- SVG colors must go through **CSS classes** (`kg-edge`, `kg-dot`,
  `kg-label` set in the SVG stylesheet, colors `var(--g-*)`) — presentation
  attributes cannot use `var()`.
- Markdown code highlighting: goldmark-highlighting + Chroma in **classes
  mode** (`chromahtml.WithClasses(true)`, prefix `chroma-`) — never inline
  colors. Token-family colors live in the inline stylesheet (`.chroma-*`)
  and MUST use the `g-syn-*` tokens (`g-syn-keyword`, `g-syn-string`,
  `g-syn-comment`, `g-syn-error`) or `g-hi` — never `g-cyan`, `g-positive`,
  `g-text` or `g-negative-text`, which do not reach 4.5:1 on `g-code`.
  Extended families (kd/kt/fm/mb/…): update the rules in `layout.go` when a
  new lexer family shows up unstyled.

### Text Colors & Contrast (WCAG AA)

- Every text/background pair used by the UI MUST reach **4.5:1**
  (`html/pages/contrast_test.go`, `TestPaletteContrast`, both themes).
- Use `g-text` for body copy, excerpts, labels, placeholders, tags,
  secondary actions and empty-state mentions. `g-muted` / `g-subdue` are
  decorative only and MUST NOT be used as a text color.
- `g-accent` and `g-negative` are decorative tokens (SVG, borders, halos,
  fills). Any **text** using those hues MUST use `g-accent-text` /
  `g-negative-text`, which are tuned per theme for contrast.
- `g-ink` is reserved for text placed on `g-accent` (primary buttons).
- Read-only form fields reuse `g-widget` + `g-text` + `cursor-not-allowed`;
  editable fields keep `g-block`.
- **Nested surfaces use their own token.** `g-widget` is the card surface;
  never reuse it as the background of an element *inside* a card (that made
  `<code>`/`<pre>` invisible). Code surfaces use `g-code` plus a
  `1px solid var(--g-border)` outline; code text is `g-hi` (`g-text` falls
  under 4.5:1 on `g-code`). `html/pages/contrast_test.go`
  (`TestCodeSurfaceDistinctFromWidget`) guards against the regression.

### Form Errors: htmx Swap vs Native Fallback

htmx does **not** swap the target of a `4xx`/`5xx` response. Therefore:

- Request carries `HX-Request: true` → respond **200 + the message node
  only** so `#reponse-serveur` is replaced and the error is visible.
- Native POST (no JavaScript) → respond the **full pre-filled form page**
  with the validation message and the proper status (`400` invalid,
  `409` duplicate). Values already typed are preserved through
  `pages.FormulaireKB{Valeurs, Message}` and `PageKBCreateWith` /
  `PageKBEditWith`.
- The faulty field is flagged with the `.field-error` class (outline
  `var(--g-negative-text)`), removed on the first `input` event. Never
  inject an inline `#dc322f` outline: colors must follow the active theme.
- Labels are bound to controls (`for="field-<nom>"` / matching `id`), and a
  control carries exactly **one** `class` attribute (append modifiers to
  the class string, never emit a second `Class(...)`).

