package authkit_test

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"io"
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
	if pgStore.TableName() != "custom_users" {
		t.Errorf("expected custom_users, got %s", pgStore.TableName())
	}

	// Schema init check with nil db returns error
	if err := pgStore.InitSchema(context.Background()); err == nil {
		t.Errorf("expected error when db is nil")
	}

	// Table name validation
	if err := authkit.ValidateTableName("users"); err != nil {
		t.Errorf("expected valid table name: %v", err)
	}
	if err := authkit.ValidateTableName("users; DROP TABLE users;"); err == nil {
		t.Errorf("expected error for SQL injection payload in table name")
	}
}

// -------------------------------------------------------------------------
// 10. Security & Hardening Tests (DoS Body Limit, OTP Cooldown, Revocation)
// -------------------------------------------------------------------------

func TestMaxBodyLimit_PayloadTooLarge(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	// Create payload larger than MaxBodyBytes (64KB default)
	largeData := make([]byte, 70*1024)
	for i := range largeData {
		largeData[i] = 'A'
	}
	largePayload, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": string(largeData),
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(largePayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 StatusRequestEntityTooLarge for oversized payload, got %d", w.Code)
	}
}

func TestOTP_Cooldown(t *testing.T) {
	engine, store, _ := setupTestEngine(t)
	handler := engine.Routes()

	email := "cooldown@example.com"
	_ = store.Create(context.Background(), &authkit.User{
		ID:           "cool_1",
		Email:        email,
		PasswordHash: "dummy",
	})

	// Request 1: Should succeed
	body1, _ := json.Marshal(authkit.ForgotPasswordRequest{Email: email})
	req1 := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("request 1: expected 200 OK, got %d: %s", w1.Code, w1.Body.String())
	}

	// Request 2 immediately within cooldown (60s): Should return 429 Too Many Requests
	body2, _ := json.Marshal(authkit.ForgotPasswordRequest{Email: email})
	req2 := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("request 2: expected 429 Too Many Requests due to cooldown, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestLogout_And_TokenRevocation(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	// Register user
	regBody, _ := json.Marshal(authkit.RegisterRequest{
		Email:    "logout_user@example.com",
		Password: "SecurePassword123!",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)

	var authResp authkit.AuthResponse
	_ = json.Unmarshal(regW.Body.Bytes(), &authResp)

	// Call /logout with refresh token
	logoutBody, _ := json.Marshal(authkit.LogoutRequest{
		RefreshToken: authResp.RefreshToken,
	})
	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewReader(logoutBody))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutW := httptest.NewRecorder()
	handler.ServeHTTP(logoutW, logoutReq)

	if logoutW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for logout, got %d", logoutW.Code)
	}

	// Attempting to refresh using the revoked token must FAIL
	refreshBody, _ := json.Marshal(authkit.RefreshTokenRequest{
		RefreshToken: authResp.RefreshToken,
	})
	refreshReq := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(refreshBody))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshW := httptest.NewRecorder()
	handler.ServeHTTP(refreshW, refreshReq)

	if refreshW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for revoked refresh token, got %d: %s", refreshW.Code, refreshW.Body.String())
	}
}

func TestRefreshToken_SingleUseRotation(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	handler := engine.Routes()

	// Register
	regBody, _ := json.Marshal(authkit.RegisterRequest{
		Email:    "rotation@example.com",
		Password: "SecurePassword123!",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)

	var authResp authkit.AuthResponse
	_ = json.Unmarshal(regW.Body.Bytes(), &authResp)

	// Refresh 1: Should SUCCEED
	refreshBody, _ := json.Marshal(authkit.RefreshTokenRequest{
		RefreshToken: authResp.RefreshToken,
	})
	refreshReq := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(refreshBody))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshW := httptest.NewRecorder()
	handler.ServeHTTP(refreshW, refreshReq)

	if refreshW.Code != http.StatusOK {
		t.Fatalf("refresh 1: expected 200 OK, got %d: %s", refreshW.Code, refreshW.Body.String())
	}

	// Refresh 2 with same OLD token (Replay attack): Should FAIL
	refreshReq2 := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(refreshBody))
	refreshReq2.Header.Set("Content-Type", "application/json")
	refreshW2 := httptest.NewRecorder()
	handler.ServeHTTP(refreshW2, refreshReq2)

	if refreshW2.Code != http.StatusUnauthorized {
		t.Errorf("refresh 2 (replay): expected 401 Unauthorized, got %d: %s", refreshW2.Code, refreshW2.Body.String())
	}
}

func TestFunctionalOptions_And_NewConfig(t *testing.T) {
	cfg := authkit.NewConfig(
		"secret-key-that-is-at-least-32-bytes-long",
		authkit.WithAccessTokenTTL(10*time.Minute),
		authkit.WithRefreshTokenTTL(48*time.Hour),
		authkit.WithBcryptCost(11),
		authkit.WithOTPTTL(3*time.Minute),
		authkit.WithOTPMaxAttempts(5),
		authkit.WithOTPCooldown(30*time.Second),
		authkit.WithMaxBodyBytes(128*1024),
		authkit.WithMinPasswordLength(10),
		authkit.WithIssuer("custom-issuer"),
		authkit.WithDefaultRole("member"),
	)

	if cfg.AccessTokenTTL != 10*time.Minute {
		t.Errorf("expected 10m AccessTokenTTL")
	}
	if cfg.RefreshTokenTTL != 48*time.Hour {
		t.Errorf("expected 48h RefreshTokenTTL")
	}
	if cfg.BcryptCost != 11 {
		t.Errorf("expected BcryptCost 11")
	}
	if cfg.OTPTTL != 3*time.Minute {
		t.Errorf("expected OTPTTL 3m")
	}
	if cfg.OTPMaxAttempts != 5 {
		t.Errorf("expected OTPMaxAttempts 5")
	}
	if cfg.OTPCooldown != 30*time.Second {
		t.Errorf("expected OTPCooldown 30s")
	}
	if cfg.MaxBodyBytes != 128*1024 {
		t.Errorf("expected MaxBodyBytes 128KB")
	}
	if cfg.MinPasswordLength != 10 {
		t.Errorf("expected MinPasswordLength 10")
	}
	if cfg.Issuer != "custom-issuer" {
		t.Errorf("expected custom-issuer")
	}
	if cfg.DefaultRole != "member" {
		t.Errorf("expected member")
	}
}

// -------------------------------------------------------------------------
// 11. Zero-Dependency Mock SQL Driver for PostgresUserStore Testing
// -------------------------------------------------------------------------

type mockDriver struct{}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{}, nil
}

type mockConn struct{}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{query: query}, nil
}
func (c *mockConn) Close() error { return nil }
func (c *mockConn) Begin() (driver.Tx, error) { return &mockTx{}, nil }

type mockTx struct{}

func (t *mockTx) Commit() error   { return nil }
func (t *mockTx) Rollback() error { return nil }

type mockStmt struct {
	query string
}

func (s *mockStmt) Close() error { return nil }
func (s *mockStmt) NumInput() int { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &mockRows{index: 0}, nil
}

type mockRows struct {
	index int
}

func (r *mockRows) Columns() []string {
	return []string{"id", "email", "password_hash", "role", "created_at", "updated_at"}
}
func (r *mockRows) Close() error { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.index >= 1 {
		return io.EOF
	}
	r.index++
	now := time.Now()
	dest[0] = "usr_123"
	dest[1] = "mock@example.com"
	dest[2] = "$2a$12$dummyhash"
	dest[3] = "user"
	dest[4] = now
	dest[5] = now
	return nil
}

func init() {
	sql.Register("mock_postgres", &mockDriver{})
}

func TestPostgresUserStore_Operations(t *testing.T) {
	db, err := sql.Open("mock_postgres", "")
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	store := authkit.NewPostgresUserStore(db, &authkit.PostgresOptions{TableName: "users"})

	ctx := context.Background()

	// InitSchema
	if err := store.InitSchema(ctx); err != nil {
		t.Errorf("InitSchema failed: %v", err)
	}

	// Create
	testUser := &authkit.User{
		Email:        "mock@example.com",
		PasswordHash: "secret_hash",
		Role:         "user",
	}
	if err := store.Create(ctx, testUser); err != nil {
		t.Errorf("Create failed: %v", err)
	}

	// GetByEmail
	u, err := store.GetByEmail(ctx, "mock@example.com")
	if err != nil {
		t.Errorf("GetByEmail failed: %v", err)
	}
	if u == nil || u.Email != "mock@example.com" {
		t.Errorf("expected user with mock@example.com")
	}

	// GetByID
	uID, err := store.GetByID(ctx, "usr_123")
	if err != nil {
		t.Errorf("GetByID failed: %v", err)
	}
	if uID == nil || uID.ID != "usr_123" {
		t.Errorf("expected user with usr_123")
	}

	// UpdatePassword
	if err := store.UpdatePassword(ctx, "usr_123", "new_hash"); err != nil {
		t.Errorf("UpdatePassword failed: %v", err)
	}
}

func TestEngine_SetRevocationStore(t *testing.T) {
	engine, _, _ := setupTestEngine(t)
	customRev := authkit.NewMemoryRevocationStore()
	engine.SetRevocationStore(customRev)
	if engine.Revocation() != customRev {
		t.Errorf("expected custom revocation store")
	}
}

// -------------------------------------------------------------------------
// 12. Benchmarks
// -------------------------------------------------------------------------

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = authkit.HashPassword("TestPassword123!", 10)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	hash, _ := authkit.HashPassword("TestPassword123!", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = authkit.CheckPassword("TestPassword123!", hash)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	store := authkit.NewMemoryUserStore()
	cfg := authkit.DefaultConfig("super-secret-key-that-is-at-least-32-bytes-long!")
	engine, _ := authkit.New(store, cfg)
	defer engine.Close()

	user := &authkit.User{ID: "usr_1", Email: "test@example.com", Role: "user"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GenerateToken(user, "access", 15*time.Minute)
	}
}

func BenchmarkGenerateCodeOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = authkit.GenerateCode()
	}
}



