package authkit

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Engine is the central core of go-authkit.
// Isme auth logic, token management, password hashing, OTP handling, and routing encapsulate hote hain.
type Engine struct {
	store      UserStore
	cfg        Config
	otp        *OTPManager
	revocation RevocationStore
	dummyHash  string
}

// New creates and initializes a ready-to-use AuthKit Engine instance.
func New(store UserStore, cfg Config) (*Engine, error) {
	if store == nil {
		return nil, errors.New("authkit: store must not be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	otpMgr := NewOTPManager(cfg.OTPTTL, cfg.OTPMaxAttempts, cfg.OTPCooldown)
	revStore := NewMemoryRevocationStore()
	dummyHash := GenerateDummyHash(cfg.BcryptCost)

	return &Engine{
		store:      store,
		cfg:        cfg,
		otp:        otpMgr,
		revocation: revStore,
		dummyHash:  dummyHash,
	}, nil
}

// SetRevocationStore swaps the default in-memory token revocation backend with a custom store (e.g. Redis).
func (e *Engine) SetRevocationStore(store RevocationStore) {
	if store != nil {
		if e.revocation != nil {
			_ = e.revocation.Close()
		}
		e.revocation = store
	}
}

// Routes returns a pre-wired chi.Mux HTTP handler with all standard auth endpoints ready.
// Yeh standard http.Handler interface satisfy karta hai, toh kisi bhi router (Chi, net/http, Gorilla)
// ke sath seamlessly plug and play ho jata hai:
//
//	r.Mount("/api/v1/auth", authEngine.Routes())
func (e *Engine) Routes() http.Handler {
	r := chi.NewRouter()

	// Public auth endpoints
	r.Post("/register", e.HandleRegister)
	r.Post("/login", e.HandleLogin)
	r.Post("/refresh", e.HandleRefresh)
	r.Post("/logout", e.HandleLogout)
	r.Post("/forgot-password", e.HandleForgotPassword)
	r.Post("/reset-password", e.HandleResetPassword)

	// Protected endpoint
	r.With(e.RequireAuth).Get("/me", e.HandleMe)

	return r
}

// Mount registers go-authkit endpoints directly onto an existing chi.Router instance.
func (e *Engine) Mount(r chi.Router) {
	r.Post("/register", e.HandleRegister)
	r.Post("/login", e.HandleLogin)
	r.Post("/refresh", e.HandleRefresh)
	r.Post("/logout", e.HandleLogout)
	r.Post("/forgot-password", e.HandleForgotPassword)
	r.Post("/reset-password", e.HandleResetPassword)
	r.With(e.RequireAuth).Get("/me", e.HandleMe)
}

// Store returns the underlying UserStore interface.
func (e *Engine) Store() UserStore {
	return e.store
}

// Config returns active runtime configuration.
func (e *Engine) Config() Config {
	return e.cfg
}

// OTP returns the underlying OTP engine.
func (e *Engine) OTP() *OTPManager {
	return e.otp
}

// Revocation returns the active token revocation store.
func (e *Engine) Revocation() RevocationStore {
	return e.revocation
}

// Close gracefully stops all background tasks (such as OTP & Revocation cleanup goroutines).
// Satisfies standard io.Closer interface.
func (e *Engine) Close() error {
	if e.otp != nil {
		_ = e.otp.Close()
	}
	if e.revocation != nil {
		_ = e.revocation.Close()
	}
	return nil
}
