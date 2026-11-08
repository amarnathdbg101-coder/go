# SPEC: go-authkit

## 1. Scope & Purpose
Production-grade, zero-boilerplate Go authentication & user security library.
- In Scope: JWT dual-token pair generation (Access + Refresh), Bcrypt password hashing (Cost 12), Crypto 6-digit OTP engine with 3-attempt brute-force lockout, RBAC middleware, and pluggable persistence (PostgreSQL + Memory).
- Out of Scope: Third-party OAuth2 social logins (Google/Facebook) - handles core authentication & identity persistence.

## 2. Public API Contract
- `New(store UserStore, cfg Config) (*Engine, error)`
- `Engine.Routes() http.Handler`
- `Engine.RequireAuth(next http.Handler) http.Handler`
- `RequireRole(allowedRoles ...string) func(http.Handler) http.Handler`
- `UserStore` Interface: `Create`, `GetByEmail`, `GetByID`, `UpdatePassword`
- `MemoryUserStore`: In-memory thread-safe user store for zero-setup local dev & testing
- `PostgresUserStore`: Production PostgreSQL user store backed by standard `database/sql`

## 3. Concurrency & Memory Guarantees
- Thread-safe storage with `sync.RWMutex`.
- OTPManager runs an automated background cleanup ticker goroutine to evict expired OTPs.
- Graceful teardown via `Engine.Close()` which closes background goroutines cleanly.

## 4. How to Extend
- Implement `UserStore` interface for new database drivers (MySQL, SQLite, MongoDB).
- Plug custom OTP dispatchers via `Config.OnSendOTP(ctx, email, otp)`.
