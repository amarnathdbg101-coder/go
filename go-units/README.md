# go-units

Human-readable byte size and extended duration (days, weeks) parser and formatter for Go.

---

## Features
- **Data Size Parsing & Formatting**: Parses `"10MB"`, `"512KiB"`, `"1.5GB"`, `"64k"`. Formats bytes into clean binary IEC (`GiB`) or decimal SI (`GB`) representations.
- **Extended Durations**: Unlike standard `time.ParseDuration`, `go-units` natively supports days (`d`) and weeks (`w`) (e.g. `"2w 3d 4h"`).
- **Configuration Friendly**: Ideal for parsing YAML/JSON configs (`max_file_size: 25MB`, `cache_ttl: 7d`).
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"
	"time"

	gounits "go-units"
)

func main() {
	// 1. Parse byte sizes from config
	maxBytes, _ := gounits.ParseBytes("50MB")
	fmt.Printf("50MB = %d bytes\n", maxBytes) // 50,000,000

	// 2. Format storage usage
	var usedBytes int64 = 1572864000
	fmt.Println("Used (IEC):", gounits.FormatBytes(usedBytes))        // 1.46 GiB
	fmt.Println("Used (SI): ", gounits.FormatBytesDecimal(usedBytes)) // 1.57 GB

	// 3. Extended duration parsing (days and weeks)
	retention, _ := gounits.ParseHumanDuration("2w 3d")
	fmt.Printf("Retention: %v (%.1f days)\n", retention, retention.Hours()/24.0)

	// 4. Format duration
	d := (48*time.Hour + 30*time.Minute)
	fmt.Println("Formatted:", gounits.FormatHumanDuration(d)) // "2d 30m"
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
