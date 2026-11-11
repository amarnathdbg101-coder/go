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

// GenerateDummyHash creates a valid Bcrypt hash used for constant-time dummy comparisons
// during login attempts with non-existent emails, defeating timing side-channel attacks.
func GenerateDummyHash(cost int) string {
	hash, err := HashPassword("anti-timing-dummy-secret-hash", cost)
	if err != nil {
		// Fallback to minimal valid bcrypt structure if generation fails
		return "$2a$12$e8kZ2aF5c84iA7c40d99uu4yJjCjY53o4c/qMhN1gG1wO7L4M7K0i"
	}
	return hash
}
