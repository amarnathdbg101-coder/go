package gojwt

import (
	"encoding/json"
	"time"
)

// Claims represents standard and custom JWT claims.
type Claims struct {
	Subject   string         `json:"sub,omitempty"`
	Issuer    string         `json:"iss,omitempty"`
	Audience  string         `json:"aud,omitempty"`
	ExpiresAt int64          `json:"exp,omitempty"`
	NotBefore int64          `json:"nbf,omitempty"`
	IssuedAt  int64          `json:"iat,omitempty"`
	ID        string         `json:"jti,omitempty"`
	Custom    map[string]any `json:"custom,omitempty"`
}

// Valid verifies expiration and not-before timestamps with optional clock skew leeway.
func (c *Claims) Valid(leeway time.Duration) error {
	now := time.Now().Unix()
	leewaySec := int64(leeway.Seconds())

	if c.ExpiresAt > 0 && now-leewaySec > c.ExpiresAt {
		return ErrTokenExpired
	}

	if c.NotBefore > 0 && now+leewaySec < c.NotBefore {
		return ErrTokenNotValidYet
	}

	return nil
}

// ToMap serializes claims into a map.
func (c *Claims) ToMap() (map[string]any, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	err = json.Unmarshal(data, &res)
	return res, err
}
