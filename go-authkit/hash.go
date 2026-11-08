package authkit

import "golang.org/x/crypto/bcrypt"

// HashPassword generates a Bcrypt cryptographic hash of the plain-text password.
// Cost factor 12 is recommended for modern hardware to prevent GPU-accelerated rainbow table attacks.
func HashPassword(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plaintext password with a Bcrypt hash in constant time.
// Returns true if password matches, false otherwise (preventing timing attacks).
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
