package hpack_test

import (
	"testing"

	"github.com/amarnathdbg101-coder/go/go-hpack"
)

func TestHPACK_StaticTableIndexed(t *testing.T) {
	enc := hpack.NewEncoder(4096)
	dec := hpack.NewDecoder(4096)

	headers := []hpack.HeaderField{
		{Name: ":method", Value: "GET"},   // Static index 2 -> 0x82 (1 byte!)
		{Name: ":scheme", Value: "https"}, // Static index 7 -> 0x87 (1 byte!)
		{Name: ":path", Value: "/"},       // Static index 4 -> 0x84 (1 byte!)
		{Name: ":status", Value: "200"},   // Static index 8 -> 0x88 (1 byte!)
	}

	compressed := enc.Encode(headers)
	// 4 headers should compress to exactly 4 bytes!
	if len(compressed) != 4 {
		t.Fatalf("expected 4 bytes compressed output, got %d bytes: %x", len(compressed), compressed)
	}

	decoded, err := dec.Decode(compressed)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !hpack.HeaderListsEqual(headers, decoded) {
		t.Fatalf("decoded mismatch: expected %v, got %v", headers, decoded)
	}
}

func TestHPACK_DynamicTableIndexing(t *testing.T) {
	enc := hpack.NewEncoder(4096)
	dec := hpack.NewDecoder(4096)

	customHeaders := []hpack.HeaderField{
		{Name: "custom-key", Value: "custom-value-12345"},
	}

	// 1st request: Literal with incremental indexing
	comp1 := enc.Encode(customHeaders)
	dec1, err := dec.Decode(comp1)
	if err != nil {
		t.Fatalf("1st decode failed: %v", err)
	}
	if !hpack.HeaderListsEqual(customHeaders, dec1) {
		t.Fatalf("1st decoded mismatch")
	}

	// 2nd request: Header is now in dynamic table! Should encode as indexed reference
	comp2 := enc.Encode(customHeaders)
	if len(comp2) >= len(comp1) {
		t.Fatalf("expected 2nd compression to be smaller due to dynamic indexing: 1st=%d, 2nd=%d", len(comp1), len(comp2))
	}

	dec2, err := dec.Decode(comp2)
	if err != nil {
		t.Fatalf("2nd decode failed: %v", err)
	}
	if !hpack.HeaderListsEqual(customHeaders, dec2) {
		t.Fatalf("2nd decoded mismatch")
	}
}

func TestHPACK_SensitiveNeverIndexed(t *testing.T) {
	enc := hpack.NewEncoder(4096)
	dec := hpack.NewDecoder(4096)

	headers := []hpack.HeaderField{
		{Name: "authorization", Value: "Bearer secret_jwt_token", Sensitive: true},
	}

	compressed := enc.Encode(headers)
	decoded, err := dec.Decode(compressed)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(decoded) != 1 || !decoded[0].Sensitive {
		t.Fatalf("expected sensitive header flag preserved, got: %v", decoded)
	}

	// Encoding again should NOT use dynamic table index because it was never indexed
	comp2 := enc.Encode(headers)
	if len(comp2) != len(compressed) {
		t.Fatalf("never indexed header should not change length on 2nd encode: 1st=%d, 2nd=%d", len(compressed), len(comp2))
	}
}

func TestHPACK_FullRequestRoundTrip(t *testing.T) {
	enc := hpack.NewEncoder(4096)
	dec := hpack.NewDecoder(4096)

	request := []hpack.HeaderField{
		{Name: ":method", Value: "POST"},
		{Name: ":scheme", Value: "https"},
		{Name: ":path", Value: "/api/v1/orders"},
		{Name: ":authority", Value: "api.shopsilo.com"},
		{Name: "content-type", Value: "application/json"},
		{Name: "user-agent", Value: "ShopSilo-Client/2.0 (Android; Go-Backend)"},
		{Name: "x-request-id", Value: "req-9876543210"},
	}

	compressed := enc.Encode(request)
	decompressed, err := dec.Decode(compressed)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !hpack.HeaderListsEqual(request, decompressed) {
		t.Fatalf("roundtrip mismatch: expected %v, got %v", request, decompressed)
	}
}

