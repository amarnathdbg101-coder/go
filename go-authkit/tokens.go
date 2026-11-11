package authkit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a signed JWT string for a specific token type ("access" or "refresh").
func (e *Engine) GenerateToken(user *User, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	tokenID := generateUniqueID()

	claims := Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    e.cfg.Issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(e.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("authkit: failed to sign token: %w", err)
	}
	return signed, nil
}

// GenerateTokenPair generates both Access Token and Refresh Token for an authenticated User.
func (e *Engine) GenerateTokenPair(user *User) (*AuthResponse, error) {
	accessToken, err := e.GenerateToken(user, "access", e.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := e.GenerateToken(user, "refresh", e.cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(e.cfg.AccessTokenTTL.Seconds()),
		User:         user.ToResponse(),
	}, nil
}

// ValidateToken parses, verifies HMAC signature, checks expiry, validates revocation, and ensures expected token type.
func (e *Engine) ValidateToken(tokenStr string, expectedType string) (*Claims, error) {
	return e.ValidateTokenWithContext(context.Background(), tokenStr, expectedType)
}

// ValidateTokenWithContext parses and verifies JWT while checking against the revocation store using the provided context.
func (e *Engine) ValidateTokenWithContext(ctx context.Context, tokenStr string, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Enforce HMAC algorithm to prevent alg: "none" or asymmetric key confusion attacks
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(e.cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Token type validation (access vs refresh isolation)
	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	// Check if this token was revoked (Logout or rotated Refresh Token)
	if e.revocation != nil && claims.ID != "" {
		revoked, revErr := e.revocation.IsRevoked(ctx, claims.ID)
		if revErr == nil && revoked {
			return nil, ErrTokenRevoked
		}
	}

	return claims, nil
}
