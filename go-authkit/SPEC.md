# SPEC: go-authkit

## 1. Scope & Purpose
Production-grade, zero-boilerplate Go authentication & user security library.
- **In Scope**: JWT dual-token pair generation (Access + Refresh), Single-use Refresh Token Rotation, Session Revocation (`POST /logout`), Constant-time Bcrypt password hashing (Cost 12) with Anti-Timing-Attack dummy verification, Crypto 6-digit OTP engine with 60s cooldown and brute-force lockout, DoS request body limits (64KB), RBAC middleware, and pluggable persistence (PostgreSQL + Memory).
- **Out of Scope**: Third-party OAuth2 social logins (Google/GitHub/Facebook) - dedicated to core authentication & identity persistence.

## 2. Public API Contract
- `New(store UserStore, cfg Config) (*Engine, error)`
- `NewConfig(jwtSecret string, opts ...Option) Config`
- `Engine.Routes() http.Handler` (Exposes `/register`, `/login`, `/refresh`, `/logout`, `/forgot-password`, `/reset-password`, `/me`)
- `Engine.Mount(r chi.Router)`
- `Engine.RequireAuth(next http.Handler) http.Handler`
- `RequireRole(allowedRoles ...string) func(http.Handler) http.Handler`
- `Engine.SetRevocationStore(store RevocationStore)`
- `Engine.Close() error` (Implements `io.Closer`)
- `UserStore` Interface: `Create`, `GetByEmail`, `GetByID`, `UpdatePassword`
- `RevocationStore` Interface: `Revoke`, `IsRevoked`, `Close`
- `MemoryUserStore`: In-memory thread-safe user store for zero-setup local dev & testing
- `MemoryRevocationStore`: In-memory thread-safe token revocation store with automatic TTL sweep
- `PostgresUserStore`: Production PostgreSQL user store backed by standard `database/sql`
- `PostgresUserStore.InitSchema(ctx context.Context) error`: Auto-creates PostgreSQL tables & indexes
- `ValidateTableName(name string) error`: Strict identifier sanitizer preventing SQL injection

## 3. Concurrency & Memory Guarantees
- Thread-safe storage with `sync.RWMutex`.
- OTPManager runs an automated background cleanup ticker goroutine to evict expired OTPs.
- MemoryRevocationStore runs an automated background cleanup ticker goroutine to evict expired JTIs.
- Anti-leak concurrency: Graceful teardown via `Engine.Close() error` closes all background routines cleanly.
- Request payload protection: Enforces `http.MaxBytesReader` (default 64KB) on all handlers.

## 4. How to Extend
- Implement `UserStore` interface for new database drivers (MySQL, SQLite, MongoDB).
- Implement `RevocationStore` interface for distributed multi-node clusters (e.g. Redis).
- Plug custom OTP dispatchers via `Config.OnSendOTP(ctx, email, otp)`.
