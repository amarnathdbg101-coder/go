# go-sysinfo 📊

`go-sysinfo` ek lightweight, zero-dependency real-time diagnostics aur telemetry collection library hai Go services ke liye. Ye process metadata, CPU cores, active goroutines, heap allocation statistics, aur garbage collection pauses ko instant point-in-time snapshot me collect karta hai for healthchecks (`/healthz`, `/readyz`), debug endpoints, aur monitoring pipelines.

---

## ✨ Features

- **Process & Host Telemetry**: Hostname, PID, OS, Architecture, CPU core count, Go runtime version, Uptime.
- **Go Memory Allocator Telemetry**: Heap Alloc, Sys, Heap In-Use, Heap Idle, Heap Released, and GC Pause latency.
- **Human-Readable Byte Units**: Built-in `FormatBytes` producing clean `MiB` and `GiB` strings.
- **Ready for JSON / REST API**: Direct `ToMap()` utility for JSON endpoints.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-sysinfo
```

### Complete Healthcheck Endpoint Example
```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/shopsilo/go-sysinfo"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	snapshot := sysinfo.Collect()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot.ToMap())
}

func main() {
	snap := sysinfo.Collect()
	fmt.Printf("Running on %s (%s/%s) with %d CPUs\n", snap.Hostname, snap.OS, snap.Arch, snap.NumCPU)
	fmt.Printf("Active Goroutines: %d | Memory Alloc: %s\n", snap.Goroutines, snap.HumanMemory()["alloc"])
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestCollect
--- PASS: TestCollect (0.00s)
=== RUN   TestFormatBytes
--- PASS: TestFormatBytes (0.00s)
=== RUN   TestHumanMemoryAndToMap
--- PASS: TestHumanMemoryAndToMap (0.00s)
PASS
ok  	go-sysinfo	0.491s
```

---

## 📄 License
MIT License.
