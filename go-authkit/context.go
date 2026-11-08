package authkit

import "context"

// contextKey is an unexported type to prevent key collisions across packages in context.WithValue.
type contextKey struct{}

var claimsKey = contextKey{}

// WithClaims injects authenticated JWT Claims into the request context.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// GetClaims retrieves JWT Claims from context if present.
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// GetUserID extracts the authenticated user ID string from context.
func GetUserID(ctx context.Context) (string, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.UserID, true
}

// GetUserEmail extracts the authenticated user's email from context.
func GetUserEmail(ctx context.Context) (string, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.Email, true
}

// GetUserRole extracts the user's role (e.g. "admin", "user") from context.
func GetUserRole(ctx context.Context) (string, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.Role, true
}
