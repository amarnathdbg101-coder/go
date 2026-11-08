package gotemplatekit

import (
	"errors"
	"strings"
	"testing"
)

type UserContext struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func TestTemplate_MapAndFilters(t *testing.T) {
	engine := New(true)

	data := map[string]any{
		"name": "amar",
		"otp":  892341,
	}

	tmpl := "Hello {{ name | upper }}, your OTP is {{ otp }}."
	out, err := engine.Render(tmpl, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := "Hello AMAR, your OTP is 892341."
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestTemplate_StructData(t *testing.T) {
	engine := New(false)

	user := UserContext{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Role:  "Admin",
	}

	tmpl := "User: {{ name }} (Role: {{ role | lower }})"
	out, err := engine.Render(tmpl, user)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := "User: Jane Doe (Role: admin)"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestTemplate_DefaultFilter(t *testing.T) {
	engine := New(false) // Lenient mode

	data := map[string]any{
		"user": "Alice",
	}

	tmpl := "Welcome {{ user }}! Status: {{ status | default:'ACTIVE' }}"
	out, err := engine.Render(tmpl, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := "Welcome Alice! Status: ACTIVE"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestTemplate_StrictModeMissingVariable(t *testing.T) {
	engine := New(true) // Strict mode

	data := map[string]any{
		"user": "Alice",
	}

	tmpl := "Hello {{ user }}, your balance is {{ balance }}"
	_, err := engine.Render(tmpl, data)
	if !errors.Is(err, ErrMissingVariable) {
		t.Errorf("expected ErrMissingVariable in strict mode, got %v", err)
	}
}

func TestTemplate_CustomFilter(t *testing.T) {
	engine := New(true)

	engine.RegisterFilter("currencysymbol", func(in, arg string) string {
		return "₹" + in
	})

	data := map[string]any{"price": "1499"}
	out, err := engine.Render("Total: {{ price | currencysymbol }}", data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "₹1499") {
		t.Errorf("expected '₹1499', got %s", out)
	}
}
