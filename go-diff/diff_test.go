package godiff

import (
	"testing"
)

type Address struct {
	City    string `json:"city"`
	Pincode string `json:"pincode"`
}

type User struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Email    string  `diff:"user_email"`
	Password string  `diff:"-"` // Must be ignored
	Address  Address `json:"address"`
	Tags     []string `json:"tags"`
}

func TestCompare_Structs(t *testing.T) {
	u1 := User{
		ID:       1,
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret-old-password",
		Address: Address{
			City:    "Mumbai",
			Pincode: "400001",
		},
		Tags: []string{"admin", "editor"},
	}

	u2 := User{
		ID:       1,
		Name:     "Alice Cooper",       // Modified
		Email:    "alice.c@example.com", // Modified (uses diff tag)
		Password: "secret-new-password", // Ignored due to diff:"-"
		Address: Address{
			City:    "Bengaluru", // Modified nested
			Pincode: "400001",
		},
		Tags: []string{"admin", "editor", "moderator"}, // Added element
	}

	res, err := Compare(u1, u2)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if !res.HasChanges() {
		t.Fatalf("expected changes, found none")
	}

	changes := make(map[string]DiffItem)
	for _, c := range res.Changes {
		changes[c.Path] = c
	}

	// Verify Name change
	if c, ok := changes["name"]; !ok || c.Type != Modified || c.OldValue != "Alice" || c.NewValue != "Alice Cooper" {
		t.Errorf("name change mismatch: %+v", c)
	}

	// Verify Email change (diff tag custom name)
	if c, ok := changes["user_email"]; !ok || c.Type != Modified || c.NewValue != "alice.c@example.com" {
		t.Errorf("user_email change mismatch: %+v", c)
	}

	// Verify Password was ignored
	if _, ok := changes["Password"]; ok {
		t.Errorf("password should have been ignored due to diff:\"-\"")
	}

	// Verify Nested Address City
	if c, ok := changes["address.city"]; !ok || c.Type != Modified || c.NewValue != "Bengaluru" {
		t.Errorf("address.city change mismatch: %+v", c)
	}

	// Verify Slice Addition
	if c, ok := changes["tags[2]"]; !ok || c.Type != Added || c.NewValue != "moderator" {
		t.Errorf("tags slice addition mismatch: %+v", c)
	}
}

func TestCompare_Maps(t *testing.T) {
	m1 := map[string]any{
		"theme": "dark",
		"items": 10,
		"beta":  true,
	}

	m2 := map[string]any{
		"theme": "light",     // Modified
		"beta":  true,        // Unchanged
		"lang":  "en-US",     // Added
		// "items" was removed
	}

	res, err := Compare(m1, m2)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if len(res.Modified()) != 1 || res.Modified()[0].Path != "theme" {
		t.Errorf("expected 1 modified change on 'theme', got: %+v", res.Modified())
	}

	if len(res.Added()) != 1 || res.Added()[0].Path != "lang" {
		t.Errorf("expected 1 added change on 'lang', got: %+v", res.Added())
	}

	if len(res.Removed()) != 1 || res.Removed()[0].Path != "items" {
		t.Errorf("expected 1 removed change on 'items', got: %+v", res.Removed())
	}
}

func TestCompare_Identical(t *testing.T) {
	a := 42
	b := 42

	res, err := Compare(a, b)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if res.HasChanges() {
		t.Errorf("expected no changes for identical values, got: %+v", res.Changes)
	}
}
