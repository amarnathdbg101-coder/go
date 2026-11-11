# go-env/envconfig ⚙️

> This focused package is part of the `go-env` module. Import it from `github.com/amarnathdbg101-coder/go/go-env/envconfig`; install the module with `go get github.com/amarnathdbg101-coder/go/go-env@v1.0.0`.

`go-env/envconfig` ek lightweight, zero-dependency environment variable unmarshaler hai Go structures ke liye. Ye 12-Factor App design pattern ko follow karta hai aur environment variables ko strongly-typed structs me safely parse karta hai using declarative struct tags.

---

## ✨ Features

- **Declarative Struct Tags**: `env:"KEY"`, `default:"VAL"`, `required:"true"`, `split:","`, `mask:"true"`.
- **Rich Type Support**: `string`, `int/8/16/32/64`, `uint/8/16/32/64`, `float32/64`, `bool`, `time.Duration`, aur slices (e.g. `[]string`, `[]int`).
- **Flexible Boolean Parsing**: `true`, `1`, `yes`, `on` -> `true` | `false`, `0`, `no`, `off` -> `false`.
- **Nested Struct Hierarchies**: Sub-structs ko prefix ke saath recursively unmarshal karta hai.
- **Leak-Proof Masked Dumps**: `DumpMasked()` se sensitive credentials (passwords, tokens, API keys) `******` ban kar log hoti hain.
- **Pluggable Lookup for Testing**: `ProcessWithLookup()` se mock map pass karke unit tests ko isolated rakhein bina OS environment pollute kiye.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/amarnathdbg101-coder/go/go-env@v1.0.0
```

### Complete Example
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/amarnathdbg101-coder/go/go-env/envconfig"
)

type DatabaseConfig struct {
	Host     string `env:"HOST" default:"localhost"`
	Port     int    `env:"PORT" default:"5432"`
	Password string `env:"PASSWORD" required:"true" mask:"true"`
}

type ServerConfig struct {
	AppName string         `env:"NAME" default:"OrderService"`
	Port    int            `env:"PORT" default:"8080"`
	Timeout time.Duration  `env:"TIMEOUT" default:"15s"`
	Cors    []string       `env:"CORS_ORIGINS" split:","`
	DB      DatabaseConfig `env:"DB"`
}

func main() {
	var cfg ServerConfig

	// Reads environment variables prefixed with "APP_"
	// e.g. APP_PORT, APP_DB_HOST, APP_DB_PASSWORD
	if err := envconfig.Process("APP", &cfg); err != nil {
		log.Fatalf("Config load failed: %v", err)
	}

	fmt.Printf("Started %s on port %d with DB host %s\n", cfg.AppName, cfg.Port, cfg.DB.Host)

	// Safe masked dump for logs
	masked := envconfig.DumpMasked(&cfg)
	fmt.Printf("Startup Config: %+v\n", masked)
	// APP_DB_PASSWORD will be "******"
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestProcess_Success
--- PASS: TestProcess_Success (0.00s)
=== RUN   TestProcess_Defaults
--- PASS: TestProcess_Defaults (0.00s)
=== RUN   TestProcess_RequiredMissing
--- PASS: TestProcess_RequiredMissing (0.00s)
=== RUN   TestProcess_NotPointerToStruct
--- PASS: TestProcess_NotPointerToStruct (0.00s)
=== RUN   TestProcess_InvalidParsing
--- PASS: TestProcess_InvalidParsing (0.00s)
=== RUN   TestDumpMasked
--- PASS: TestDumpMasked (0.00s)
PASS
ok  	github.com/amarnathdbg101-coder/go/go-env/envconfig	0.480s
```

---

## 📄 License
MIT License.
