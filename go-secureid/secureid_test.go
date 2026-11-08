package gosecureid

import (
	"strings"
	"testing"
)

func TestToken(t *testing.T) {
	tok, err := Token(32)
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}

	if len(tok) == 0 {
		t.Errorf("empty token returned")
	}

	// Should not contain padding '=' or standard base64 '+' / '/'
	if strings.ContainsAny(tok, "=+/") {
		t.Errorf("token contains non-URL-safe characters: %s", tok)
	}

	// Invalid length check
	_, err = Token(0)
	if err != ErrInvalidLength {
		t.Errorf("expected ErrInvalidLength, got %v", err)
	}
}

func TestAPIKey(t *testing.T) {
	key, err := APIKey("sk_live", 24)
	if err != nil {
		t.Fatalf("APIKey failed: %v", err)
	}

	if !strings.HasPrefix(key, "sk_live_") {
		t.Errorf("expected prefix 'sk_live_', got %s", key)
	}
}

func TestNanoID(t *testing.T) {
	id, err := NanoID(21)
	if err != nil {
		t.Fatalf("NanoID failed: %v", err)
	}

	if len(id) != 21 {
		t.Errorf("expected len 21, got %d", len(id))
	}

	for _, c := range id {
		if !strings.ContainsRune(nanoIDAlphabet, c) {
			t.Errorf("character %c not in NanoID alphabet", c)
		}
	}
}

func TestInviteCode(t *testing.T) {
	code, err := InviteCode(8)
	if err != nil {
		t.Fatalf("InviteCode failed: %v", err)
	}

	if len(code) != 8 {
		t.Errorf("expected len 8, got %d", len(code))
	}

	// Crockford Base32 must NOT contain ambiguous characters: I, L, O, U
	if strings.ContainsAny(code, "ILOU") {
		t.Errorf("invite code contains ambiguous character: %s", code)
	}
}

func TestUniqueness_NoCollisions(t *testing.T) {
	seen := make(map[string]bool)
	total := 1000

	for i := 0; i < total; i++ {
		id, err := NanoID(16)
		if err != nil {
			t.Fatalf("NanoID failed: %v", err)
		}
		if seen[id] {
			t.Fatalf("collision detected at iteration %d: %s", i, id)
		}
		seen[id] = true
	}
}
