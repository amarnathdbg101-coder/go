# go-useragent 🌐

`go-useragent` ek ultra-fast, zero-dependency HTTP `User-Agent` parser library hai Go ke liye. Ye HTTP requests se Browser, Browser Version, Operating System, Device Family (Desktop, Mobile, Tablet, Bot), aur Web Crawlers ko sub-microsecond speed me detect karta hai bina kisi heavy regex engine ke.

---

## ✨ Features

- **Blazing Fast Detection**: Regex overhead ke bina pre-indexed token resolution.
- **Accurate Browser Classification**: Priority order ensures Edge, Opera, Samsung Internet, Chrome, Firefox, aur Safari bina collision ke correctly detect hote hain.
- **Granular Device Types**: `Desktop`, `Mobile`, `Tablet`, `Bot`, `Unknown`.
- **Bot / Crawler Detection**: Googlebot, Bingbot, curl, Postman, WhatsApp, FacebookBot automatically detect hote hain.
- **HTTP Handler Integration**: Direct `FromRequest(r)` utility for standard `http.HandlerFunc` or Chi/Gin middlewares.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-useragent
```

### Complete HTTP Middleware Example
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/shopsilo/go-useragent"
)

func auditLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := useragent.FromRequest(r)

		if ua.IsBot {
			fmt.Printf("[BOT DETECTED] Path: %s\n", r.URL.Path)
		} else {
			fmt.Printf("[USER SESSION] Browser: %s %s | OS: %s | Device: %s\n",
				ua.Browser, ua.BrowserVersion, ua.OS, ua.Device)
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	raw := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 Version/17.4.1 Mobile/15E148 Safari/604.1"
	ua := useragent.Parse(raw)

	fmt.Println("Browser:", ua.Browser)        // Safari
	fmt.Println("Version:", ua.BrowserVersion) // 17.4.1
	fmt.Println("OS:", ua.OS)                  // iOS
	fmt.Println("Device:", ua.Device)          // Mobile
	fmt.Println("IsMobile:", ua.IsMobile)      // true
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestParse_DesktopBrowsers
--- PASS: TestParse_DesktopBrowsers (0.00s)
=== RUN   TestParse_MobileAndTablet
--- PASS: TestParse_MobileAndTablet (0.00s)
=== RUN   TestParse_Bots
--- PASS: TestParse_Bots (0.00s)
=== RUN   TestFromRequest
--- PASS: TestFromRequest (0.00s)
PASS
ok  	go-useragent	0.639s
```

---

## 📄 License
MIT License.
