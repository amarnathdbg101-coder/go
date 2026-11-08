# go-authkit 🚀

> **High-Performance, Modular, Plug-and-Play Authentication Library for Go.**  
> Built with Go Standard Library idioms, Chi compatibility, PostgreSQL support, and zero boilerplate.

---

## ✨ Features

- ⚡ **Plug & Play Routing**: Mount complete authentication into Chi or standard `net/http` in **3 lines of code**.
- 🔐 **Dual-Token JWT Security**: 15-minute Access Tokens + 7-day Refresh Tokens with cryptographic type isolation (prevents token swap attacks).
- 🛡️ **Bcrypt Hash Security**: Cost factor 12 hashing with constant-time verification against timing attacks.
- 📱 **Crypto OTP Engine**: 6-digit numeric OTPs via `crypto/rand` with in-memory TTL, brute-force lockout (max 3 attempts), and automated background cleanup.
- 👥 **Role-Based Access Control (RBAC)**: Flexible `RequireRole("admin", "manager")` middleware.
- 🗄️ **Pluggable Storage (`UserStore`)**:
  - `PostgresUserStore`: Production-grade PostgreSQL store using Go standard library `database/sql`.
  - `MemoryUserStore`: In-memory thread-safe store for instant unit testing and local prototyping.
- 🎯 **Context Helpers**: Type-safe extraction of `GetUserID(ctx)`, `GetUserEmail(ctx)`, and `GetUserRole(ctx)`.

---

## 📦 Installation

```bash
go get github.com/shopsilo/go-authkit
```

---

## ⚡ 3-Minute Quickstart

### 1. In-Memory Mode (Zero DB Setup - Test instantly)

```go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shopsilo/go-authkit"
)

func main() {
	// 1. Initialize store & configuration
	store := authkit.NewMemoryUserStore()
	cfg := authkit.DefaultConfig("super-secret-jwt-key-must-be-at-least-16-bytes")

	// 2. Create AuthKit Engine
	authEngine, err := authkit.New(store, cfg)
	if err != nil {
		log.Fatalf("AuthKit initialization failed: %v", err)
	}
	defer authEngine.Close()

	// 3. Mount onto your Chi Router
	r := chi.NewRouter()
	r.Mount("/auth", authEngine.Routes())

	// Protected application route
	r.Group(func(r chi.Router) {
		r.Use(authEngine.RequireAuth)
		r.Get("/profile", func(w http.ResponseWriter, req *http.Request) {
			userID, _ := authkit.GetUserID(req.Context())
			w.Write([]byte("Welcome, User: " + userID))
		})
	})

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
```

---

## 🗄️ Production PostgreSQL Integration

### 1. Database Table Schema (`schema.sql`)

Execute this in `psql` or run via your migration tool:

```sql
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

### 2. Connect in Go

```go
package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shopsilo/go-authkit"
)

func main() {
	// DSN Connection String format: postgres://user:password@host:port/dbname?sslmode=disable
	dsn := "postgres://postgres:postgres@localhost:5432/myapp_db?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// Initialize Postgres store
	store := authkit.NewPostgresUserStore(db, &authkit.PostgresOptions{
		TableName: "users",
	})

	cfg := authkit.DefaultConfig("my-production-very-secure-jwt-secret-string-32")

	authEngine, err := authkit.New(store, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer authEngine.Close()

	r := chi.NewRouter()
	r.Mount("/api/v1/auth", authEngine.Routes())

	http.ListenAndServe(":8080", r)
}
```

---

## 📡 Included HTTP Endpoints

| Method | Route | Description | Auth Required |
|---|---|---|---|
| `POST` | `/register` | Sign up with email & password | No |
| `POST` | `/login` | Authenticate credentials & get JWT tokens | No |
| `POST` | `/refresh` | Rotate access token using refresh token | No |
| `POST` | `/forgot-password` | Generate 6-digit secure OTP | No |
| `POST` | `/reset-password` | Verify OTP & set new password | No |
| `GET` | `/me` | Get profile details of current user | **Yes** (Bearer Token) |

---

## 🛡️ Middleware & Context Helpers

### Protected Route with Auth Middleware

```go
r.With(authEngine.RequireAuth).Get("/api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
    userID, _ := authkit.GetUserID(r.Context())
    email, _  := authkit.GetUserEmail(r.Context())
    // Business logic...
})
```

### Role-Based Access Control (RBAC)

```go
// Only users with "admin" or "manager" role can access
r.With(authEngine.RequireAuth, authkit.RequireRole("admin", "manager")).
    Get("/api/v1/admin/analytics", handleAdminAnalytics)
```

---

## 📧 Sending Custom OTPs (Email / SMS Hook)

You can plug in your own email sender (Resend, SendGrid, SMTP, Twilio) cleanly via `OnSendOTP`:

```go
cfg := authkit.DefaultConfig("your-jwt-secret-key-at-least-16-bytes")
cfg.OnSendOTP = func(ctx context.Context, email, otp string) error {
    log.Printf("[EMAIL SENDER] Sending OTP %s to %s", otp, email)
    // Call your Resend / SendGrid / SMTP client here:
    // return emailClient.Send(email, "Your Verification Code", "OTP: " + otp)
    return nil
}
```

---

## 🧪 Testing

Run comprehensive unit tests:

```bash
cd C:\Library\go-authkit
go test -v ./...
```
