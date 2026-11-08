# go-netkit 🔌

`go-netkit` ek essential, zero-dependency network utilities library hai Go backends aur microservices ke liye. Ye ephemeral port allocation (`GetFreePort`), active TCP port reachability probes (`IsPortOpen`), dependency polling (`WaitForPort`), comprehensive RFC private/public IP classification, aur CIDR subnet containment check provide karta hai.

---

## ✨ Features

- **Ephemeral Port Allocation**: Dynamic test environments me race-free open TCP ports allocate karein (`GetFreePort`).
- **Active TCP Reachability Probe**: DB ya Redis reachability check karein without crashing (`IsPortOpen`).
- **Polling Dependency Waiter**: Container ya background service ready hone tak context-aware wait karein (`WaitForPort`).
- **RFC IP Classification**: RFC 1918 (10.x, 172.16.x, 192.168.x), RFC 6598 (CGNAT 100.64.x), Loopback, aur IPv6 ULA (`fc00::/7`).
- **CIDR Subnet Whitelisting**: Check whether an IP falls inside a CIDR block or slice (`ContainsIP`, `ContainsAnyIP`).
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-netkit
```

### Complete Example
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shopsilo/go-netkit"
)

func main() {
	// 1. Allocate a free port for a dynamic mock server
	freePort, err := netkit.GetFreePort()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Allocated free test port: %d\n", freePort)

	// 2. Wait up to 5 seconds for Postgres container to become ready
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := netkit.WaitForPort(ctx, "localhost", 5432, 100*time.Millisecond); err != nil {
		fmt.Printf("Postgres port not ready: %v\n", err)
	} else {
		fmt.Println("Postgres port is OPEN and accepting connections!")
	}

	// 3. Private IP & CIDR Check (Security Middleware)
	clientIP := "192.168.1.45"
	if netkit.IsPrivateIPString(clientIP) {
		fmt.Println("Request originated from an internal private network.")
	}

	// 4. CIDR Whitelist check
	isAllowed := netkit.ContainsAnyIP([]string{"10.0.0.0/8", "192.168.0.0/16"}, clientIP)
	fmt.Printf("Client IP %s allowed: %v\n", clientIP, isAllowed)
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestGetFreePort
--- PASS: TestGetFreePort (0.01s)
=== RUN   TestIsPortOpen
--- PASS: TestIsPortOpen (0.00s)
=== RUN   TestWaitForPort
--- PASS: TestWaitForPort (0.05s)
=== RUN   TestPrivateAndPublicIP
--- PASS: TestPrivateAndPublicIP (0.00s)
=== RUN   TestContainsIP
--- PASS: TestContainsIP (0.00s)
PASS
ok  	go-netkit	0.614s
```

---

## 📄 License
MIT License.
