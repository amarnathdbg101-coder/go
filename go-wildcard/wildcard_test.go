package gowildcard

import (
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern  string
		text     string
		expected bool
	}{
		{"*", "anything", true},
		{"", "", true},
		{"exact", "exact", true},
		{"exact", "different", false},
		{"*.png", "image.png", true},
		{"*.png", "image.jpeg", false},
		{"api/*/users", "api/v1/users", true},
		{"api/*/users", "api/v1/v2/users", true},
		{"file?.txt", "file1.txt", true},
		{"file?.txt", "file10.txt", false},
		{"*order*paid*", "webhook.order.item.paid.event", true},
		{"*order*paid*", "webhook.order.item.failed.event", false},
		{"sub.*.domain.com", "sub.eu.domain.com", true},
	}

	for _, tt := range tests {
		actual := Match(tt.pattern, tt.text)
		if actual != tt.expected {
			t.Errorf("Match(%q, %q): expected %v, got %v", tt.pattern, tt.text, tt.expected, actual)
		}
	}
}

func TestMatchFold(t *testing.T) {
	if !MatchFold("*.PNG", "photo.png") {
		t.Errorf("expected *.PNG to match photo.png with MatchFold")
	}
	if !MatchFold("USER_?_EVENT", "user_1_event") {
		t.Errorf("expected USER_?_EVENT to match user_1_event with MatchFold")
	}
}

func TestFilter(t *testing.T) {
	items := []string{
		"config.yaml",
		"main.go",
		"README.md",
		"utils.go",
		"server.go",
	}

	goFiles := Filter("*.go", items)
	expected := []string{"main.go", "utils.go", "server.go"}

	if !reflect.DeepEqual(goFiles, expected) {
		t.Errorf("Filter expected %v, got %v", expected, goFiles)
	}
}

func BenchmarkMatch(b *testing.B) {
	pattern := "api/v*/users/*/profile"
	text := "api/v2/users/usr_98124/profile"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Match(pattern, text)
	}
}
