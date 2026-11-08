# Technical Specification: go-netkit

## 1. Domain & Purpose
`go-netkit` is a lightweight, zero-dependency network utility toolkit for Go backend services and microservices. It facilitates ephemeral port allocation (`GetFreePort`), active TCP port connectivity checks (`IsPortOpen`), polling waiters (`WaitForPort`), comprehensive private/public IP classification (RFC 1918, RFC 6598, RFC 4193), and CIDR subnet containment testing.

## 2. Architecture & Design Principles
- **Ephemeral Port Allocation**: Safely binds to TCP `:0`, retrieves the kernel-assigned port, and releases the listener for race-free dynamic service testing.
- **Resilient Polling Waiter**: `WaitForPort` polls remote endpoints with configurable intervals and `context.Context` cancellation to reliably synchronize test harnesses with external dependencies (e.g. Postgres, Redis).
- **Standards-Compliant IP Classification**: Accurately classifies IPv4 and IPv6 addresses across private, loopback, link-local, and Carrier-Grade NAT (CGNAT) address spaces.
- **CIDR Subnet Whitelisting**: `ContainsIP` and `ContainsAnyIP` evaluate CIDR masks for network security perimeter validation.
- **Zero Heavy Dependencies**: Pure standard library (`net`, `time`, `context`, `strconv`, `fmt`).

## 3. API Surface
- `func GetFreePort() (int, error)`
- `func IsPortOpen(host string, port int, timeout time.Duration) bool`
- `func WaitForPort(ctx context.Context, host string, port int, interval time.Duration) error`
- `func IsPrivateIP(ip net.IP) bool`
- `func IsPrivateIPString(ipStr string) bool`
- `func IsPublicIP(ip net.IP) bool`
- `func ContainsIP(cidr string, ipStr string) (bool, error)`
- `func ContainsAnyIP(cidrs []string, ipStr string) bool`

## 4. Error Handling
- `ErrPortUnavailable`: Kernel could not assign an open ephemeral TCP socket.
- `ErrInvalidCIDR`: Malformed CIDR subnet string.
- `ErrInvalidIP`: Malformed IP address string.
- `ErrTimeout`: Context expired before target TCP port became reachable.

## 5. Verification & Testing
- Unit tests verify ephemeral port allocation, open/closed port detection, background async port waiting, private/public IP classification tables, and CIDR subnet containment.
