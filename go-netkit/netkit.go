package netkit

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

var privateCIDRs []*net.IPNet

func init() {
	blocks := []string{
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918
		"192.168.0.0/16", // RFC 1918
		"100.64.0.0/10",  // RFC 6598 CGNAT
		"127.0.0.0/8",    // RFC 1122 Loopback
		"169.254.0.0/16", // RFC 3927 Link-local
		"::1/128",        // IPv6 Loopback
		"fc00::/7",       // RFC 4193 IPv6 Unique Local Address
		"fe80::/10",      // RFC 4291 IPv6 Link-Local
	}

	for _, b := range blocks {
		_, ipNet, err := net.ParseCIDR(b)
		if err == nil {
			privateCIDRs = append(privateCIDRs, ipNet)
		}
	}
}

// GetFreePort requests an ephemeral open TCP port from the OS kernel and returns it.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrPortUnavailable, err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrPortUnavailable, err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// IsPortOpen checks whether a TCP connection can be established to host:port within timeout.
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// WaitForPort polls host:port until a TCP connection succeeds or ctx is cancelled.
func WaitForPort(ctx context.Context, host string, port int, interval time.Duration) error {
	if interval <= 0 {
		interval = 50 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial immediate probe
	if IsPortOpen(host, port, interval/2) {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		case <-ticker.C:
			if IsPortOpen(host, port, interval/2) {
				return nil
			}
		}
	}
}

// IsPrivateIP returns true if ip belongs to private, loopback, link-local, or CGNAT subnets.
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	for _, cidr := range privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// IsPrivateIPString parses an IP string and returns whether it is a private IP.
func IsPrivateIPString(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return IsPrivateIP(ip)
}

// IsPublicIP returns true if ip is valid and neither private nor loopback/multicast.
func IsPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	return !IsPrivateIP(ip)
}

// ContainsIP checks if ipStr is contained inside the subnet specified by cidr.
func ContainsIP(cidr string, ipStr string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidCIDR, err)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, ErrInvalidIP
	}

	return ipNet.Contains(ip), nil
}

// ContainsAnyIP checks if ipStr matches any of the given CIDR blocks.
func ContainsAnyIP(cidrs []string, ipStr string) bool {
	for _, c := range cidrs {
		if match, err := ContainsIP(c, ipStr); err == nil && match {
			return true
		}
	}
	return false
}
