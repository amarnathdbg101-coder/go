package netkit

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("unexpected error getting free port: %v", err)
	}

	if port <= 1024 {
		t.Errorf("expected ephemeral port > 1024, got %d", port)
	}

	// Verify we can actually bind to this port
	l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatalf("failed to bind to allocated free port: %v", err)
	}
	l.Close()
}

func TestIsPortOpen(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port

	// Open port
	if !IsPortOpen("127.0.0.1", port, 100*time.Millisecond) {
		t.Errorf("expected port %d to be open", port)
	}

	// Close listener
	l.Close()

	// Should be closed now
	if IsPortOpen("127.0.0.1", port, 50*time.Millisecond) {
		t.Errorf("expected port %d to be closed after listener stopped", port)
	}
}

func TestWaitForPort(t *testing.T) {
	freePort, _ := GetFreePort()

	// Launch listener with delay
	go func() {
		time.Sleep(30 * time.Millisecond)
		l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(freePort))
		if err == nil {
			defer l.Close()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := WaitForPort(ctx, "127.0.0.1", freePort, 15*time.Millisecond)
	if err != nil {
		t.Errorf("expected WaitForPort to succeed, got %v", err)
	}
}

func TestPrivateAndPublicIP(t *testing.T) {
	privateIPs := []string{
		"10.0.0.1",
		"172.16.5.2",
		"172.31.255.255",
		"192.168.1.1",
		"127.0.0.1",
		"100.64.0.1",   // CGNAT
		"169.254.1.1",  // Link-local
		"::1",          // IPv6 loopback
		"fc00::1",      // IPv6 ULA
	}

	for _, ipStr := range privateIPs {
		if !IsPrivateIPString(ipStr) {
			t.Errorf("expected %s to be private", ipStr)
		}
		ip := net.ParseIP(ipStr)
		if IsPublicIP(ip) {
			t.Errorf("expected %s NOT to be public", ipStr)
		}
	}

	publicIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"142.250.190.46", // Google
		"2607:f8b0:4005:805::200e",
	}

	for _, ipStr := range publicIPs {
		if IsPrivateIPString(ipStr) {
			t.Errorf("expected %s NOT to be private", ipStr)
		}
		ip := net.ParseIP(ipStr)
		if !IsPublicIP(ip) {
			t.Errorf("expected %s to be public", ipStr)
		}
	}
}

func TestContainsIP(t *testing.T) {
	match, err := ContainsIP("192.168.1.0/24", "192.168.1.55")
	if err != nil || !match {
		t.Errorf("expected match for 192.168.1.55 in subnet")
	}

	match2, err := ContainsIP("192.168.1.0/24", "192.168.2.1")
	if err != nil || match2 {
		t.Errorf("expected false for 192.168.2.1")
	}

	// ContainsAnyIP
	whitelist := []string{"10.0.0.0/8", "172.16.0.0/12"}
	if !ContainsAnyIP(whitelist, "10.5.2.1") {
		t.Errorf("expected ContainsAnyIP to return true for 10.5.2.1")
	}
	if ContainsAnyIP(whitelist, "8.8.8.8") {
		t.Errorf("expected ContainsAnyIP to return false for 8.8.8.8")
	}
}
