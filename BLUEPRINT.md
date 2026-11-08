# C:\Library - The Modular Go Architecture Blueprint 🏛️

> **The Constitution & Design System for High-Performance, Secure, and Lightweight Go Libraries.**  
> *Any human developer or AI agent creating, reading, or modifying libraries in `C:\Library` MUST adhere to this blueprint.*

---

## 🎯 Core Tenets (Non-Negotiable)

1. **Lightweight**: Prefer pure Go Standard Library (`net/http`, `sync`, `context`, `database/sql`, `image`, `os`). Avoid heavy monolithic frameworks (Fiber, Gin, GORM). Keep `go.mod` slim.
2. **Fastest / High-Performance**: Sub-millisecond execution. Pre-allocate slices, minimize heap escapes, avoid runtime reflection (`reflect`) by leveraging Go Generics.
3. **Security by Default**:
   - Prevent SQL Injection via strict whitelist sanitization.
   - Prevent Path Traversal via `filepath.Clean` and boundary checks.
   - Prevent Timing Attacks via `subtle.ConstantTimeCompare`.
4. **High Flexibility (Zero Coupling)**:
   - Never couple a library to a specific business domain (e.g. no hardcoded "ShopSilo" product/order logic).
   - Use Go Generics (`[T any]`), interfaces, and functional options (`With...`).
5. **Anti-Leak Concurrency**:
   - Every background goroutine must have a clean `stopChan` or `context.Done()` exit path.
   - Implement graceful shutdown via `Close() error`.

---

## 📂 Standard File Convention for Every Library

Every package `C:\Library\<name>` must follow this layout:

```
C:\Library\<name>\
├── errors.go       # Sentinel domain errors (ErrNotFound, ErrInvalidInput, etc.)
├── config.go       # Configuration struct, DefaultConfig(), Validate(), Functional Options
├── driver.go       # Interface contracts for pluggable backends (Storage, Store, Driver)
├── memory.go       # Built-in in-memory zero-dependency driver (Instant local dev & testing)
├── engine.go       # Core orchestration engine & domain state
├── <name>.go       # Top-level package entrypoint & 1-line ergonomic helpers
├── <name>_test.go  # Exhaustive unit tests, edge cases, and concurrency stress tests
├── README.md       # Developer documentation, 2-minute quickstart, and architecture
└── SPEC.md         # AI-SPEC: Machine-readable architecture & contracts for AI agents
```

---

## 🤖 `SPEC.md` Template (AI-to-AI Communication)

Every library must include a `SPEC.md` so any AI agent can understand, audit, and extend it without reading thousands of lines of raw code:

```markdown
# SPEC: <library-name>

## 1. Scope & Purpose
What problem this library solves and what it deliberately does NOT do.

## 2. Public API Contract
Key structs, interfaces, and primary methods.

## 3. Concurrency & Memory Guarantees
Locks (sync.RWMutex), goroutine lifecycle, memory bounds.

## 4. How to Extend
Rules for adding new storage backends, middleware, or features without breaking backward compatibility.
```

---

## 🧪 Quality Gates Before Merging Any Library

1. `go build ./...` must compile with 0 errors.
2. `go vet ./...` must pass with 0 warnings.
3. `go test -v -count=1 ./...` must pass with 100% green tests.
4. Pure standard library fallback must be provided (e.g. MemoryStore) so tests run without external servers.
