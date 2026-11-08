package authkit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/amarnathdbg101-coder/go/go-authkit"
)

func setupTestEngine(t *testing.T) (*authkit.Engine, *authkit.MemoryUserStore, *otpRecorder) {
	t.Helper()

	store := authkit.NewMemoryUserStore()
	rec := &otpRecorder{}

	cfg := authkit.DefaultConfig("super-secret-key-that-is-at-least-32-bytes-long!")
	cfg.AccessTokenTTL = 5 * time.Minute
	cfg.RefreshTokenTTL = 1 * time.Hour
	cfg.OTPTTL = 5 * time.Minute
	cfg.OTPMaxAttempts = 3
	cfg.OnSendOTP = func(ctx context.Context, email, otp string) error {
		rec.record(email, otp)
		return nil
	}

	engine, err := authkit.New(store, cfg)
	if err != nil {
		t.Fatalf("failed to initialize authkit engine: %v", err)
	}

	t.Cleanup(func() {
		engine.Close()
	})

	return engine, store, rec
}

type otpRecorder struct {
	mu   sync.Mutex
	otps map[string]string
}

func (r *otpRecorder) record(email, otp string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.otps == nil {
		r.otps = make(map[string]string)
	}
	r.otps[email] = otp
}

func (r *otpRecorder) get(email string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.otps[email]
}

// -------------------------------------------------------------------------
// 1. Registration Tests
// -------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	payload := authkit.RegisterRequest{
		Email:    "test@example.com",
		Password: "SecurePassword123!",
		Role:     "user",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp authkit.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse auth response: %v", err)
	}

	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Errorf("expected non-empty access and refresh tokens")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", resp.User.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	payload := authkit.RegisterRequest{
		Email:    "duplicate@example.com",
		Password: "Password123!",
	}
	body, _ := json.Marshal(payload)

	// First registration
	req1 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first registration failed: %d", w1.Code)
	}

	// Second registration with same email
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected status 409 Conflict, got %d", w2.Code)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	payload := authkit.RegisterRequest{
		Email:    "weak@example.com",
		Password: "short", // < 8 characters
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request, got %d", w.Code)
	}
}

// -------------------------------------------------------------------------
// 2. Login Tests
// -------------------------------------------------------------------------

func TestLogin_SuccessAndFailure(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	// Register a test user
	regPayload, _ := json.Marshal(authkit.RegisterRequest{
		Email:    "loginuser@example.com",
		Password: "MyPassword123!",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)

	// Case A: Valid credentials
	loginBody, _ := json.Marshal(authkit.LoginRequest{
		Email:    "loginuser@example.com",
		Password: "MyPassword123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp authkit.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.AccessToken == "" {
		t.Errorf("expected valid access token on login")
	}

	// Case B: Invalid password
	badLoginBody, _ := json.Marshal(authkit.LoginRequest{
		Email:    "loginuser@example.com",
		Password: "WrongPassword!",
	})
	badReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(badLoginBody))
	badReq.Header.Set("Content-Type", "application/json")
	badW := httptest.NewRecorder()
	handler.ServeHTTP(badW, badReq)

	if badW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for bad password, got %d", badW.Code)
	}
}

// -------------------------------------------------------------------------
// 3. Token Refresh Tests
// -------------------------------------------------------------------------

func TestRefreshToken_Flow(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	// Register user
	regPayload, _ := json.Marshal(authkit.RegisterRequest{
		Email:    "refresh@example.com",
		Password: "Password123!",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)

	var regResp authkit.AuthResponse
	_ = json.Unmarshal(regW.Body.Bytes(), &regResp)

	// Case A: Refresh using Refresh Token (Valid)
	refreshBody, _ := json.Marshal(authkit.RefreshTokenRequest{
		RefreshToken: regResp.RefreshToken,
	})
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on token refresh, got %d: %s", w.Code, w.Body.String())
	}

	var refreshResp authkit.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &refreshResp)
	if refreshResp.AccessToken == "" {
		t.Errorf("expected new access token")
	}

	// Case B: Attacker tries to use AccessToken as RefreshToken (Should Fail)
	badRefreshBody, _ := json.Marshal(authkit.RefreshTokenRequest{
		RefreshToken: regResp.AccessToken, // Misuse of AccessToken
	})
	badReq := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(badRefreshBody))
	badReq.Header.Set("Content-Type", "application/json")
	badW := httptest.NewRecorder()
	handler.ServeHTTP(badW, badReq)

	if badW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized when using AccessToken in refresh endpoint, got %d", badW.Code)
	}
}

// -------------------------------------------------------------------------
// 4. RequireAuth & /me Protected Endpoint Tests
// -------------------------------------------------------------------------

func TestRequireAuth_And_Me(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	// Register user
	regPayload, _ := json.Marshal(authkit.RegisterRequest{
		Email:    "me_user@example.com",
		Password: "Password123!",
		Role:     "user",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)

	var regResp authkit.AuthResponse
	_ = json.Unmarshal(regW.Body.Bytes(), &regResp)

	// Case A: No token -> 401
	unauthReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	unauthW := httptest.NewRecorder()
	handler.ServeHTTP(unauthW, unauthReq)
	if unauthW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth header, got %d", unauthW.Code)
	}

	// Case B: Valid Bearer Token -> 200 OK with UserResponse
	authReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	authReq.Header.Set("Authorization", "Bearer "+regResp.AccessToken)
	authW := httptest.NewRecorder()
	handler.ServeHTTP(authW, authReq)

	if authW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /me, got %d: %s", authW.Code, authW.Body.String())
	}

	var userResp authkit.UserResponse
	_ = json.Unmarshal(authW.Body.Bytes(), &userResp)
	if userResp.Email != "me_user@example.com" {
		t.Errorf("expected email 'me_user@example.com', got '%s'", userResp.Email)
	}
}

// -------------------------------------------------------------------------
// 5. RequireRole RBAC Middleware Tests
// -------------------------------------------------------------------------

func TestRequireRole_Middleware(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	// Setup custom Chi router with an admin-only endpoint
	r := chi.NewRouter()
	r.With(engine.RequireAuth, authkit.RequireRole("admin")).Get("/admin/dashboard", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"admin": true}`))
	})

	// 1. User with "user" role
	userToken, err := engine.GenerateToken(&authkit.User{ID: "u1", Email: "u1@test.com", Role: "user"}, "access", 10*time.Minute)
	if err != nil {
		t.Fatalf("token gen error: %v", err)
	}

	userReq := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	userReq.Header.Set("Authorization", "Bearer "+userToken)
	userW := httptest.NewRecorder()
	r.ServeHTTP(userW, userReq)

	if userW.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-admin, got %d", userW.Code)
	}

	// 2. User with "admin" role
	adminToken, err := engine.GenerateToken(&authkit.User{ID: "u2", Email: "admin@test.com", Role: "admin"}, "access", 10*time.Minute)
	if err != nil {
		t.Fatalf("token gen error: %v", err)
	}

	adminReq := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	adminReq.Header.Set("Authorization", "Bearer "+adminToken)
	adminW := httptest.NewRecorder()
	r.ServeHTTP(adminW, adminReq)

	if adminW.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin, got %d", adminW.Code)
	}
}

// -------------------------------------------------------------------------
// 6. Forgot Password, OTP Lockout & Reset Password Tests
// -------------------------------------------------------------------------

func TestForgotPassword_And_ResetPassword_Flow(t *testing.T) {
	engine, _, rec := setupTestEngine(t)
	handler := engine.Routes()

	email := "forgotuser@example.com"
	initialPassword := "OldPassword123!"

	// Register user
	regPayload, _ := json.Marshal(authkit.RegisterRequest{
		Email:    email,
		Password: initialPassword,
	})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)

	// Step 1: Request Forgot Password
	fpPayload, _ := json.Marshal(authkit.ForgotPasswordRequest{Email: email})
	fpReq := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(fpPayload))
	fpReq.Header.Set("Content-Type", "application/json")
	fpW := httptest.NewRecorder()
	handler.ServeHTTP(fpW, fpReq)

	if fpW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for forgot-password, got %d", fpW.Code)
	}

	capturedOTP := rec.get(email)
	if len(capturedOTP) != 6 {
		t.Fatalf("expected 6-digit OTP, got '%s'", capturedOTP)
	}

	// Step 2: Try invalid OTP
	wrongResetPayload, _ := json.Marshal(authkit.ResetPasswordRequest{
		Email:       email,
		OTP:         "000000", // Wrong OTP
		NewPassword: "BrandNewPassword123!",
	})
	wrongReq := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(wrongResetPayload))
	wrongReq.Header.Set("Content-Type", "application/json")
	wrongW := httptest.NewRecorder()
	handler.ServeHTTP(wrongW, wrongReq)

	if wrongW.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for wrong OTP, got %d", wrongW.Code)
	}

	// Step 3: Valid Reset Password with captured OTP
	validResetPayload, _ := json.Marshal(authkit.ResetPasswordRequest{
		Email:       email,
		OTP:         capturedOTP,
		NewPassword: "BrandNewPassword123!",
	})
	validReq := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(validResetPayload))
	validReq.Header.Set("Content-Type", "application/json")
	validW := httptest.NewRecorder()
	handler.ServeHTTP(validW, validReq)

	if validW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for password reset, got %d: %s", validW.Code, validW.Body.String())
	}

	// Step 4: Login with Old Password should FAIL
	oldLoginPayload, _ := json.Marshal(authkit.LoginRequest{
		Email:    email,
		Password: initialPassword,
	})
	oldLoginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(oldLoginPayload))
	oldLoginReq.Header.Set("Content-Type", "application/json")
	oldLoginW := httptest.NewRecorder()
	handler.ServeHTTP(oldLoginW, oldLoginReq)

	if oldLoginW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for old password after reset, got %d", oldLoginW.Code)
	}

	// Step 5: Login with New Password should SUCCEED
	newLoginPayload, _ := json.Marshal(authkit.LoginRequest{
		Email:    email,
		Password: "BrandNewPassword123!",
	})
	newLoginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(newLoginPayload))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginW := httptest.NewRecorder()
	handler.ServeHTTP(newLoginW, newLoginReq)

	if newLoginW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login with new password, got %d", newLoginW.Code)
	}
}

// -------------------------------------------------------------------------
// 7. Brute Force OTP Lockout Test
// -------------------------------------------------------------------------

func TestOTP_MaxAttemptsLockout(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	email := "lockout@example.com"
	_ = engine.Store().Create(context.Background(), &authkit.User{
		ID:           "lock_1",
		Email:        email,
		PasswordHash: "dummy",
	})

	// Request OTP
	fpPayload, _ := json.Marshal(authkit.ForgotPasswordRequest{Email: email})
	fpReq := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(fpPayload))
	fpReq.Header.Set("Content-Type", "application/json")
	fpW := httptest.NewRecorder()
	handler.ServeHTTP(fpW, fpReq)

	// Attempt 1: wrong OTP
	badPayload, _ := json.Marshal(authkit.ResetPasswordRequest{
		Email:       email,
		OTP:         "111111",
		NewPassword: "NewPassword123!",
	})
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(badPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("attempt %d: expected 400, got %d", i, w.Code)
		}
	}

	// Attempt 3: 3rd wrong attempt should trigger lockout (429 Too Many Requests)
	req3 := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(badPayload))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests on 3rd attempt, got %d", w3.Code)
	}
}

// -------------------------------------------------------------------------
// 8. Config & Constructor Edge Case Tests
// -------------------------------------------------------------------------

func TestConfig_Validation(t *testing.T) {
	// Case A: Empty secret
	cfg1 := authkit.Config{}
	if err := cfg1.Validate(); err == nil {
		t.Errorf("expected error for empty JWTSecret")
	}

	// Case B: Secret too short (< 16 chars)
	cfg2 := authkit.Config{JWTSecret: "short"}
	if err := cfg2.Validate(); err == nil {
		t.Errorf("expected error for short JWTSecret")
	}

	// Case C: Valid secret populates defaults
	cfg3 := authkit.DefaultConfig("this-is-a-valid-32-byte-secret-key-123")
	if err := cfg3.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if cfg3.AccessTokenTTL <= 0 || cfg3.RefreshTokenTTL <= 0 {
		t.Errorf("expected positive TTLs")
	}

	// Nil store check
	if _, err := authkit.New(nil, cfg3); err == nil {
		t.Errorf("expected error when store is nil")
	}
}

// -------------------------------------------------------------------------
// 9. Context & Engine Helpers Tests
// -------------------------------------------------------------------------

func TestContextHelpers_Empty(t *testing.T) {
	ctx := context.Background()

	if _, ok := authkit.GetClaims(ctx); ok {
		t.Errorf("expected false for empty claims")
	}
	if _, ok := authkit.GetUserID(ctx); ok {
		t.Errorf("expected false for empty user ID")
	}
	if _, ok := authkit.GetUserEmail(ctx); ok {
		t.Errorf("expected false for empty email")
	}
	if _, ok := authkit.GetUserRole(ctx); ok {
		t.Errorf("expected false for empty role")
	}
}

func TestEngine_Mount(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	r := chi.NewRouter()
	engine.Mount(r)

	// Ensure routes are accessible
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without token, /me returns 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", w.Code)
	}

	// Getter tests
	if engine.Store() == nil {
		t.Errorf("expected store")
	}
	if engine.OTP() == nil {
		t.Errorf("expected otp manager")
	}
	if engine.Config().JWTSecret == "" {
		t.Errorf("expected config")
	}
}

func TestPostgresUserStore_Constructor(t *testing.T) {
	pgStore := authkit.NewPostgresUserStore(nil, &authkit.PostgresOptions{TableName: "custom_users"})
	if pgStore == nil {
		t.Fatalf("expected non-nil PostgresUserStore")
	}
}

