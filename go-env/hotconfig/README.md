# go-env/hotconfig 🔄

> **Lock-Free, Zero-Downtime Hot Configuration Reloader for Go (Go 1.23+).**  
> Update application settings (ports, database pool limits, feature flags, secret tokens) without restarting your server, dropping in-flight requests, or encountering race conditions.

---

## 🚀 Key Features

1. **Sub-Nanosecond Lock-Free Reads**:
   - `cfg.Get()` uses Go 1.23+ `atomic.Pointer[T].Load()`, compiling down to a single CPU atomic read with zero mutex contention.
2. **Cryptographic SHA-256 Verification**:
   - Computes SHA-256 hash of file contents; skips redundant parsing if content has not changed.
3. **Corruption & Panic Shield**:
   - If an editor or deploy script saves a truncated or malformed file, parsing fails safely and the previous valid configuration is preserved.
4. **Schema Validation Gates**:
   - Enforce port ranges, URL schemes, and mandatory fields via `WithValidator`. Invalid configurations are rejected before swapping.
5. **Dynamic Change Callbacks**:
   - Listen for configuration transitions via `WithOnChange(func(oldCfg, newCfg *T))`.
6. **Zero Dependencies**:
   - 100% pure standard library (`sync/atomic`, `os`, `crypto/sha256`, `encoding/json`, `time`, `sync`).

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"time"
	"github.com/amarnathdbg101-coder/go/go-env/hotconfig"
)

type Config struct {
	Port       int  `json:"port"`
	EnableLogs bool `json:"enable_logs"`
}

func main() {
	// 1. Start watching config.json
	cfg, err := hotconfig.Watch[Config]("config.json",
		hotconfig.WithOnChange[Config](func(oldCfg, newCfg *Config) {
			fmt.Printf("Config swapped! Port changed from %d to %d\n", oldCfg.Port, newCfg.Port)
		}),
	)
	if err != nil {
		panic(err)
	}
	defer cfg.Close()

	// 2. Lock-free read in hot path (zero mutex overhead):
	for {
		current := cfg.Get()
		fmt.Printf("Active Port: %d (Logs: %v)\n", current.Port, current.EnableLogs)
		time.Sleep(2 * time.Second)
	}
}
```
