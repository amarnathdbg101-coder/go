package authkit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
)

// HandleRegister handles user signup (POST /register).
func (e *Engine) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || !isValidEmail(req.Email) {
		writeJSONError(w, http.StatusBadRequest, ErrInvalidEmail.Error())
		return
	}

	if len(req.Password) < 8 {
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
func (e *Engine) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
		// Constant-time dummy verify can prevent user enumeration, but ErrInvalidCredentials is standard
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
func (e *Engine) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		writeJSONError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Validate refresh token specifically
	claims, err := e.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	user, err := e.store.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Associated user account no longer exists")
		return
	}

	authResp, err := e.GenerateTokenPair(user)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to rotate authentication tokens")
		return
	}

	writeJSON(w, http.StatusOK, authResp)
}

// HandleForgotPassword triggers 6-digit OTP generation for password reset (POST /forgot-password).
func (e *Engine) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	if len(req.NewPassword) < 8 {
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
