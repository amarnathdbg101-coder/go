package authkit

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents the standard user domain model across any backend application.
// Is struct me user ka primary data store hota hai.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // JSON serialization me password hash kabhi expose nahi hoga
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserResponse is the safe, publicly shareable representation of a User.
// Sensitive fields jaise PasswordHash isme nahi hote.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts internal User model to safe UserResponse.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// Claims defines the standard JWT Claims payload with custom domain attributes.
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// RegisterRequest holds incoming signup payload.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // Default role "user" assign hoga agar empty ho
}

// LoginRequest holds incoming credentials payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshTokenRequest holds payload for rotating access token using refresh token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest holds payload for revoking a refresh token session.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ForgotPasswordRequest holds the user's email to initiate OTP generation.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest holds verified OTP along with the new password.
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

// AuthResponse holds standard token pair returned on successful login or register.
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"` // e.g. "Bearer"
	ExpiresIn    int64        `json:"expires_in"` // Access token TTL in seconds
	User         UserResponse `json:"user"`
}

// MessageResponse is a standard JSON format for general success or error messages.
type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
