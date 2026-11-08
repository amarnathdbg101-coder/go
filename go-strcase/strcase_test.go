package gostrcase

import (
	"reflect"
	"testing"
)

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"fooBar", []string{"foo", "Bar"}},
		{"FooBar", []string{"Foo", "Bar"}},
		{"foo_bar_baz", []string{"foo", "bar", "baz"}},
		{"foo-bar-baz", []string{"foo", "bar", "baz"}},
		{"HTTPServer", []string{"HTTP", "Server"}},
		{"userID", []string{"user", "ID"}},
		{"APIKey", []string{"API", "Key"}},
		{"nested.folder/path", []string{"nested", "folder", "path"}},
	}

	for _, tt := range tests {
		actual := SplitWords(tt.input)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("SplitWords(%q): expected %v, got %v", tt.input, tt.expected, actual)
		}
	}
}

func TestToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fooBar", "foo_bar"},
		{"FooBar", "foo_bar"},
		{"HTTPServer", "http_server"},
		{"userID", "user_id"},
		{"APIResponseCode", "api_response_code"},
		{"kebab-case-string", "kebab_case_string"},
		{"SCREAMING_SNAKE", "screaming_snake"},
	}

	for _, tt := range tests {
		actual := ToSnake(tt.input)
		if actual != tt.expected {
			t.Errorf("ToSnake(%q): expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo_bar", "fooBar"},
		{"HTTP_SERVER", "httpServer"},
		{"user-id", "userId"},
		{"FooBarBaz", "fooBarBaz"},
	}

	for _, tt := range tests {
		actual := ToCamel(tt.input)
		if actual != tt.expected {
			t.Errorf("ToCamel(%q): expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestToPascal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo_bar", "FooBar"},
		{"http_server", "HttpServer"},
		{"user_id", "UserId"},
		{"kebab-case", "KebabCase"},
	}

	for _, tt := range tests {
		actual := ToPascal(tt.input)
		if actual != tt.expected {
			t.Errorf("ToPascal(%q): expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fooBar", "foo-bar"},
		{"HTTPServer", "http-server"},
		{"user_id", "user-id"},
	}

	for _, tt := range tests {
		actual := ToKebab(tt.input)
		if actual != tt.expected {
			t.Errorf("ToKebab(%q): expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestToScreamingSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"userId", "USER_ID"},
		{"httpServer", "HTTP_SERVER"},
		{"kebab-slug", "KEBAB_SLUG"},
	}

	for _, tt := range tests {
		actual := ToScreamingSnake(tt.input)
		if actual != tt.expected {
			t.Errorf("ToScreamingSnake(%q): expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func BenchmarkToSnake(b *testing.B) {
	s := "HTTPServerConfigurationOption"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToSnake(s)
	}
}
