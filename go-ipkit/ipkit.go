package goipkit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

var defaultPrivateCIDRs = []string{
	"10.0.0.0/8",     // RFC 1918 Class A
	"172.16.0.0/12",  // RFC 1918 Class B
	"192.168.0.0/16", // RFC 1918 Class C
	"127.0.0.0/8",    // IPv4 Loopback
	"169.254.0.0/16", // Link local
	"::1/128",        // IPv6 Loopback
	"fc00::/7",       // Unique local address
	"fe80::/10",      // Link local unicast
}

// Extractor safely identifies the end-user client IP behind reverse proxies and CDNs.
type Extractor struct {
	trustedSubnets []*net.IPNet
	trustPrivate   bool
}

// Option configures Extractor behavior.
type Option func(*Extractor)

// WithTrustedCIDRs adds subnets (e.g. "10.0.0.0/8", "192.168.1.0/24") whose proxy headers are trusted.
func WithTrustedCIDRs(cidrs ...string) Option {
	return func(e *Extractor) {
		for _, c := range cidrs {
			_, subnet, err := net.ParseCIDR(strings.TrimSpace(c))
			if err == nil && subnet != nil {
				e.trustedSubnets = append(e.trustedSubnets, subnet)
			}
		}
	}
}

// WithTrustPrivate enables trusting all RFC 1918 private subnets (standard for AWS VPC / K8s internal networks).
func WithTrustPrivate(trust bool) Option {
	return func(e *Extractor) {
		e.trustPrivate = trust
	}
}

// New constructs an Extractor.
func New(opts ...Option) *Extractor {
	e := &Extractor{
		trustPrivate: true, // Default to trusting internal VPC load balancers
	}

	for _, opt := range opts {
		opt(e)
	}

	if e.trustPrivate {
		for _, c := range defaultPrivateCIDRs {
			_, subnet, err := net.ParseCIDR(c)
			if err == nil && subnet != nil {
				e.trustedSubnets = append(e.trustedSubnets, subnet)
			}
		}
	}

	return e
}

// Default is a standard production extractor trusting private network load balancers.
var Default = New()

// GetRealIP extracts the real client IP address from the HTTP request using trusted proxy backward traversal.
func (e *Extractor) GetRealIP(r *http.Request) string {
	remoteIPStr := CleanIP(r.RemoteAddr)
	remoteIP := net.ParseIP(remoteIPStr)
	if remoteIP == nil {
		return remoteIPStr
	}

	// If remote address is NOT a trusted proxy, it must be the real client directly
	if !e.isTrusted(remoteIP) {
		return remoteIPStr
	}

	// Remote address is trusted: check CDN headers
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		if ip := net.ParseIP(cfIP); ip != nil {
			return cfIP
		}
	}
	if trueClientIP := strings.TrimSpace(r.Header.Get("True-Client-IP")); trueClientIP != "" {
		if ip := net.ParseIP(trueClientIP); ip != nil {
			return trueClientIP
		}
	}

	// Traverse X-Forwarded-For backwards
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		// Walk backwards from the proxy side towards the client
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			ip := net.ParseIP(candidate)
			if ip == nil {
				continue
			}

			// First non-trusted IP in the chain is the actual external client
			if !e.isTrusted(ip) {
				return candidate
			}
		}

		// If all were trusted, return the leftmost original caller
		if len(parts) > 0 {
			leftmost := strings.TrimSpace(parts[0])
			if net.ParseIP(leftmost) != nil {
				return leftmost
			}
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	return remoteIPStr
}

func (e *Extractor) isTrusted(ip net.IP) bool {
	for _, subnet := range e.trustedSubnets {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

// CleanIP strips port numbers and enclosing brackets from IPv4 or IPv6 strings.
func CleanIP(addr string) string {
	addr = strings.TrimSpace(addr)
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		return strings.Trim(host, "[]")
	}
	return strings.Trim(addr, "[]")
}

// IsPrivate checks if an IP belongs to private or local loopback subnets.
func IsPrivate(ipStr string) bool {
	ip := net.ParseIP(CleanIP(ipStr))
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	return false
}

// InCIDR checks if an IP is within a specific CIDR subnet.
func InCIDR(ipStr, cidr string) (bool, error) {
	ip := net.ParseIP(CleanIP(ipStr))
	if ip == nil {
		return false, fmt.Errorf("ipkit: invalid ip: %s", ipStr)
	}

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("ipkit: invalid cidr %s: %w", cidr, err)
	}

	return subnet.Contains(ip), nil
}
