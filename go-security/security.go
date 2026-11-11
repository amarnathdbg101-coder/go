package security

import (
	"github.com/shopsilo/go-security/auth"
)

// Version of the unified go-security suite.
const Version = "1.0.0"

// Re-exported types for top-level convenience.
type User = auth.User
type AuthConfig = auth.Config
type AuthEngine = auth.Engine

// NewAuth initializes a production authentication engine.
func NewAuth(store auth.UserStore, cfg auth.Config) (*auth.Engine, error) {
	return auth.New(store, cfg)
}
