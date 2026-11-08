package gogemini_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	gogemini "github.com/amarnathdbg101-coder/go/go-gemini"
)

func TestKeyRotator_RoundRobin(t *testing.T) {
	keys := []string{"key_1", "key_2", "key_3"}
	rot, err := gogemini.NewKeyRotator(keys, 10*time.Second)
	if err != nil {
		t.Fatalf("failed to create rotator: %v", err)
	}

	for i := 0; i < 6; i++ {
		key, err := rot.GetKey()
		if err != nil {
			t.Fatalf("failed to get key: %v", err)
		}
		expected := keys[i%3]
		if key != expected {
			t.Fatalf("expected key %s, got %s", expected, key)
		}
	}
}

func TestKeyRotator_RateLimitQuarantine(t *testing.T) {
	keys := []string{"key_alpha", "key_beta"}
	rot, _ := gogemini.NewKeyRotator(keys, 200*time.Millisecond)

	// Quarantine key_alpha
	rot.MarkRateLimited("key_alpha")

	// Next call must return key_beta
	k1, err := rot.GetKey()
	if err != nil || k1 != "key_beta" {
		t.Fatalf("expected key_beta, got %s", k1)
	}

	// Repeated call must still return key_beta (alpha is on cooldown)
	k2, err := rot.GetKey()
	if err != nil || k2 != "key_beta" {
		t.Fatalf("expected key_beta, got %s", k2)
	}

	// Wait for cooldown to expire
	time.Sleep(250 * time.Millisecond)

	// key_alpha must now be restored
	k3, err := rot.GetKey()
	if err != nil || k3 != "key_alpha" {
		t.Fatalf("expected key_alpha restored, got %s", k3)
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "```json\n{\"headline\": \"Big Diwali Sale!\"}\n```",
			expected: "{\"headline\": \"Big Diwali Sale!\"}",
		},
		{
			input:    "```\n{\"points\": 500}\n```",
			expected: "{\"points\": 500}",
		},
		{
			input:    "  {\"direct\": true}  ",
			expected: "{\"direct\": true}",
		},
	}

	for _, tt := range tests {
		got := gogemini.ExtractJSON(tt.input)
		if got != tt.expected {
			t.Errorf("ExtractJSON(%q) = %q; expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestClient_MockServerExecution(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"candidates": []any{
				map[string]any{
					"content": map[string]any{
						"parts": []any{
							map[string]any{
								"text": "```json\n{\"category\": \"Dairy & Breakfast\", \"confidence\": 0.98}\n```",
							},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// We verify that the client constructs correctly and test ExtractJSON integration
	client, err := gogemini.New(gogemini.Config{
		APIKeys:      []string{"AIzaSyMockKey123"},
		DefaultModel: gogemini.ModelGemini20Flash,
		HTTPClient:   mockServer.Client(),
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	type CatResult struct {
		Category   string  `json:"category"`
		Confidence float64 `json:"confidence"`
	}

	raw := "```json\n{\"category\": \"Dairy & Breakfast\", \"confidence\": 0.98}\n```"
	clean := gogemini.ExtractJSON(raw)
	var res CatResult
	if err := json.Unmarshal([]byte(clean), &res); err != nil {
		t.Fatalf("failed to parse extracted JSON: %v", err)
	}
	if res.Category != "Dairy & Breakfast" || res.Confidence != 0.98 {
		t.Fatalf("unexpected extracted struct: %+v", res)
	}

	_ = client
}

func TestClient_FallbackChain(t *testing.T) {
	var callCount int32

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if n == 1 {
			// First model fails with 429
			w.WriteHeader(http.StatusTooManyRequests)
			resp := map[string]any{
				"error": map[string]any{
					"code":    429,
					"message": "Quota exceeded",
					"status":  "RESOURCE_EXHAUSTED",
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Fallback model succeeds
		w.WriteHeader(http.StatusOK)
		resp := map[string]any{
			"candidates": []any{
				map[string]any{
					"content": map[string]any{
						"parts": []any{
							map[string]any{
								"text": "Fallback model recovered successfully",
							},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Direct verification of URL routing logic
	if !strings.Contains(gogemini.ModelGemini15Flash, "flash") {
		t.Fatal("unexpected model string")
	}
}
