package contenttype

import (
	"errors"
	"testing"
)

func TestParseContentType(t *testing.T) {
	header := "application/json; charset=utf-8; boundary=xyz123"
	mt, err := ParseContentType(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mt.MIME != "application/json" {
		t.Errorf("expected application/json, got %s", mt.MIME)
	}
	if mt.Params["charset"] != "utf-8" {
		t.Errorf("expected charset utf-8, got %s", mt.Params["charset"])
	}
	if mt.Params["boundary"] != "xyz123" {
		t.Errorf("expected boundary xyz123, got %s", mt.Params["boundary"])
	}
}

func TestParseAccept(t *testing.T) {
	header := "text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8"
	accepted := ParseAccept(header)

	if len(accepted) != 5 {
		t.Fatalf("expected 5 media types, got %d", len(accepted))
	}

	// Highest quality (1.0) specific types first
	if accepted[0].Quality != 1.0 {
		t.Errorf("expected first item to have quality 1.0, got %f", accepted[0].Quality)
	}

	// Lowest quality should be */* with q=0.8
	last := accepted[len(accepted)-1]
	if last.MIME != "*/*" || last.Quality != 0.8 {
		t.Errorf("expected last item to be */* with q=0.8, got %+v", last)
	}
}

func TestNegotiate(t *testing.T) {
	supported := []string{"application/json", "application/xml", "text/plain"}

	// Direct match
	match, ok := Negotiate("application/xml, application/json;q=0.5", supported)
	if !ok || match != "application/xml" {
		t.Errorf("expected application/xml, got %s (ok=%v)", match, ok)
	}

	// Wildcard subtype
	match2, ok := Negotiate("application/*", supported)
	if !ok || match2 != "application/json" {
		t.Errorf("expected application/json, got %s", match2)
	}

	// Global wildcard
	match3, ok := Negotiate("*/*", supported)
	if !ok || match3 != "application/json" {
		t.Errorf("expected application/json, got %s", match3)
	}

	// Unsupported
	_, ok = Negotiate("image/png", supported)
	if ok {
		t.Errorf("expected false for unsupported media type")
	}
}

func TestNegotiateLanguage(t *testing.T) {
	supported := []string{"en-US", "hi-IN", "es-ES"}

	match, ok := NegotiateLanguage("hi-IN, en-US;q=0.8", supported)
	if !ok || match != "hi-IN" {
		t.Errorf("expected hi-IN, got %s (ok=%v)", match, ok)
	}

	// Prefix matching: "hi" matches "hi-IN"
	match2, ok := NegotiateLanguage("hi;q=0.9", supported)
	if !ok || match2 != "hi-IN" {
		t.Errorf("expected hi-IN from 'hi', got %s", match2)
	}
}

func TestNegotiateEncoding(t *testing.T) {
	supported := []string{"br", "gzip"}

	match, ok := NegotiateEncoding("gzip, deflate, br", supported)
	if !ok || match != "br" {
		t.Errorf("expected br, got %s (ok=%v)", match, ok)
	}
}

func TestErrors(t *testing.T) {
	_, err := ParseContentType("")
	if !errors.Is(err, ErrEmptyHeader) {
		t.Errorf("expected ErrEmptyHeader, got %v", err)
	}

	_, err = ParseContentType("invalid-no-slash")
	if !errors.Is(err, ErrInvalidMediaType) {
		t.Errorf("expected ErrInvalidMediaType, got %v", err)
	}
}
