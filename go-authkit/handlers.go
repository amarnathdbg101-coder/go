package authkit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
)

// readJSON decodes incoming request body while enforcing memory limits against DoS payload exhaustion attacks.
func readJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, dst interface{}) error {
	if maxBytes <= 0 {
		maxBytes = 64 * 1024 // 64 KB default
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return ErrPayloadTooLarge
		}
		return err
	}
	return nil
}

// HandleRegister handles user signup (POST /register).
func (e *Engine) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := readJSON(w, r, e.cfg.MaxBodyBytes, &req); err != nil {
		if errors.Is(err, ErrPayloadTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || !isValidEmail(req.Email) {
		writeJSONError(w, http.StatusBadRequest, ErrInvalidEmail.Error())
		return
	}

	if len(req.Password) < e.cfg.MinPasswordLength {
		writeJSONError(w, http.StatusBadRequest, ErrWeakPassword.Error())
		return
	}

	// Check if already registered
	existing, err := e.store.GetByEmail(r.Context(), req.Email)
	if err == nil && existing != nil {
		writeJSONError(w, http.StatusConflict, ErrEmailAlreadyExists.Error())
		return
	}

	// Hash password using Bcrypt
	hash, err := HashPassword(req.Password, e.cfg.BcryptCost)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = e.cfg.DefaultRole
	}

	newUser := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         role,
	}

	if err := e.store.Create(r.Context(), newUser); err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			writeJSONError(w, http.StatusConflict, ErrEmailAlreadyExists.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to create user account")
		return
	}

	// Generate JWT Access & Refresh Token Pair
	authResp, err := e.GenerateTokenPair(newUser)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate authentication tokens")
		return
	}

	writeJSON(w, http.StatusCreated, authResp)
}

// HandleLogin authenticates credentials and returns JWT token pair (POST /login).
// Constant-time execution is guaranteed even if user does not exist, neutralizing user enumeration timing attacks.
func (e *Engine) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := readJSON(w, r, e.cfg.MaxBodyBytes, &req); err != nil {
		if errors.Is(err, ErrPayloadTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := e.store.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		// Mitigate timing attack: execute Bcrypt against dummy hash to match real verify latency
		_ = CheckPassword(req.Password, e.dummyHash)
		writeJSONError(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
		return
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		writeJSONError(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
		return
	}

	authResp, err := e.GenerateTokenPair(user)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate authentication tokens")
		return
	}

	writeJSON(w, http.StatusOK, authResp)
}

// HandleRefresh rotates access token using a valid Refresh Token (POST /refresh).
// Implements Refresh Token Rotation: previous refresh token is immediately revoked upon use.
func (e *Engine) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := readJSON(w, r, e.cfg.MaxBodyBytes, &req); err != nil {
		if errors.Is(err, ErrPayloadTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		writeJSONError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Validate refresh token specifically
	claims, err := e.ValidateTokenWithContext(r.Context(), req.RefreshToken, "refresh")
	if err != nil {
		if errors.Is(err, ErrTokenRevoked) {
			writeJSONError(w, http.StatusUnauthorized, "Refresh token has been revoked")
			return
		}
		writeJSONError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	user, err := e.store.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Associated user account no longer exists")
		return
	}

	// Refresh Token Rotation: revoke the old token JTI to prevent replay attacks
	if claims.ID != "" && e.revocation != nil {
		_ = e.revocation.Revoke(r.Context(), claims.ID, claims.ExpiresAt.Time)
	}

	authResp, err := e.GenerateTokenPair(user)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to rotate authentication tokens")
		return
	}

	writeJSON(w, http.StatusOK, authResp)
}

// HandleLogout revokes the provided refresh token session (POST /logout).
func (e *Engine) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := readJSON(w, r, e.cfg.MaxBodyBytes, &req); err != nil {
		if errors.Is(err, ErrPayloadTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken != "" {
		claims, err := e.ValidateTokenWithContext(r.Context(), req.RefreshToken, "refresh")
		if err == nil && claims.ID != "" && e.revocation != nil {
			_ = e.revocation.Revoke(r.Context(), claims.ID, claims.ExpiresAt.Time)
		}
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// HandleForgotPassword triggers 6-digit OTP generation for password reset (POST /forgot-password).
func (e *Engine) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := readJSON(w, r, e.cfg.MaxBodyBytes, &req); err != nil {
		if errors.Is(err, ErrPayloadTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !isValidEmail(req.Email) {
		writeJSONError(w, http.StatusBadRequest, ErrInvalidEmail.Error())
		return
	}

	// Check if user exists
	user, err := e.store.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		// Generic success message to prevent user enumeration attacks in production
		writeJSON(w, http.StatusOK, MessageResponse{
			Success: true,
			Message: "If your email is registered, you will receive an OTP shortly.",
		})
		return
	}

	otp, err := e.otp.Create(req.Email)
	if err != nil {
		if errors.Is(err, ErrOTPCooldown) {
			writeJSONError(w, http.StatusTooManyRequests, ErrOTPCooldown.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate OTP")
		return
	}

	// Trigger external sender hook if configured
	if e.cfg.OnSendOTP != nil {
		if err := e.cfg.OnSendOTP(r.Context(), req.Email, otp); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to dispatch OTP notification")
			return
		}
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		Success: true,
		Message: "OTP sent successfully. Please check your email.",
	})
}

// HandleResetPassword verifies OTP and updates password (POST /reset-password).
func (e *Engine) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := readJSON(w, r, e.cfg.MaxBodyBytes, &req); err != nil {
		if errors.Is(err, ErrPayloadTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.OTP = strings.TrimSpace(req.OTP)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Email == "" || req.OTP == "" {
		writeJSONError(w, http.StatusBadRequest, "Email and OTP are required")
		return
	}

	if len(req.NewPassword) < e.cfg.MinPasswordLength {
		writeJSONError(w, http.StatusBadRequest, ErrWeakPassword.Error())
		return
	}

	// Verify OTP
	if err := e.otp.Verify(req.Email, req.OTP); err != nil {
		if errors.Is(err, ErrOTPExpired) {
			writeJSONError(w, http.StatusBadRequest, ErrOTPExpired.Error())
			return
		}
		if errors.Is(err, ErrMaxAttemptsExceeded) {
			writeJSONError(w, http.StatusTooManyRequests, ErrMaxAttemptsExceeded.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, ErrInvalidOTP.Error())
		return
	}

	// Fetch user
	user, err := e.store.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, ErrUserNotFound.Error())
		return
	}

	// Hash new password
	hash, err := HashPassword(req.NewPassword, e.cfg.BcryptCost)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to hash new password")
		return
	}

	// Update password in persistence
	if err := e.store.UpdatePassword(r.Context(), user.ID, hash); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		Success: true,
		Message: "Password updated successfully. You can now login with your new password.",
	})
}

// HandleMe returns the currently authenticated user's profile (GET /me).
// Protected by RequireAuth middleware.
func (e *Engine) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := e.store.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, ErrUserNotFound.Error())
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

// writeJSON serializes response payload as JSON.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// isValidEmail checks RFC 5322 email syntax using standard library net/mail.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
