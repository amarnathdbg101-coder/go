package goipkit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractor_DirectConnection(t *testing.T) {
	extractor := New(WithTrustPrivate(true))

	// Direct client connection from public IP (e.g. 203.0.113.50)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.50:54123"

	ip := extractor.GetRealIP(req)
	if ip != "203.0.113.50" {
		t.Fatalf("expected direct IP 203.0.113.50, got %s", ip)
	}
}

func TestExtractor_BehindPrivateProxy(t *testing.T) {
	extractor := New(WithTrustPrivate(true))

	// AWS ALB / Nginx at 10.0.1.20 forwarding real client 198.51.100.99
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.1.20:8080"
	req.Header.Set("X-Forwarded-For", "198.51.100.99, 10.0.1.20")

	ip := extractor.GetRealIP(req)
	if ip != "198.51.100.99" {
		t.Fatalf("expected client IP 198.51.100.99, got %s", ip)
	}
}

func TestExtractor_AntiSpoofing(t *testing.T) {
	extractor := New(WithTrustPrivate(true))

	// Attacker hits the server directly from public IP 185.220.101.5
	// Attacker pretends to be 8.8.8.8 via fake X-Forwarded-For header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "185.220.101.5:41234"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")

	ip := extractor.GetRealIP(req)
	// Because 185.220.101.5 is NOT a trusted proxy, X-Forwarded-For must be ignored!
	if ip != "185.220.101.5" {
		t.Fatalf("anti-spoofing failed! Attacker spoofed IP: got %s, expected %s", ip, "185.220.101.5")
	}
}

func TestExtractor_Cloudflare(t *testing.T) {
	extractor := New(WithTrustPrivate(true))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:80" // Internal ingress
	req.Header.Set("CF-Connecting-IP", "103.21.244.2")

	ip := extractor.GetRealIP(req)
	if ip != "103.21.244.2" {
		t.Fatalf("expected CF-Connecting-IP, got %s", ip)
	}
}

func TestCleanIP(t *testing.T) {
	if CleanIP("192.168.1.1:8080") != "192.168.1.1" {
		t.Fatal("failed to clean IPv4 port")
	}
	if CleanIP("[2001:db8::1]:443") != "2001:db8::1" {
		t.Fatal("failed to clean IPv6 port")
	}
	if CleanIP("10.0.0.1") != "10.0.0.1" {
		t.Fatal("failed plain IP")
	}
}

func TestIsPrivate(t *testing.T) {
	if !IsPrivate("192.168.1.100") {
		t.Fatal("192.168.1.100 should be private")
	}
	if !IsPrivate("10.50.0.1") {
		t.Fatal("10.50.0.1 should be private")
	}
	if !IsPrivate("127.0.0.1") {
		t.Fatal("127.0.0.1 should be private (loopback)")
	}
	if IsPrivate("8.8.8.8") {
		t.Fatal("8.8.8.8 should NOT be private")
	}
}

func TestInCIDR(t *testing.T) {
	match, err := InCIDR("10.20.30.40", "10.0.0.0/8")
	if err != nil || !match {
		t.Fatalf("expected IP to match CIDR, err: %v", err)
	}

	match, err = InCIDR("172.30.0.1", "192.168.0.0/16")
	if err != nil || match {
		t.Fatal("expected IP to NOT match CIDR")
	}
}
