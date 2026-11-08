# go-contenttype 📋

`go-contenttype` ek lightweight HTTP content negotiation aur MIME media-type parsing library hai Go ke liye. Ye HTTP APIs aur gateways ke liye `Content-Type`, `Accept`, `Accept-Language`, aur `Accept-Encoding` headers ko quality factor ($q$-values) aur specificity order ke sath negotiate karta hai.

---

## ✨ Features

- **Media Type & Parameter Parsing**: Extract base MIME type, charset, and parameters (`application/json; charset=utf-8`).
- **Quality Factor ($q$-value) Sorting**: Correctly prioritizes client preferences (`text/html;q=0.9, application/json;q=1.0`).
- **Multi-Level Matching**: Supports exact match, wildcard subtype (`application/*`), and global wildcard (`*/*`).
- **Language & Encoding Negotiation**: Regional language prefix fallback (`hi` -> `hi-IN`) and compression negotiation (`br`, `gzip`).
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-contenttype
```

### Complete Example
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/shopsilo/go-contenttype"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Content Negotiation
	supportedFormats := []string{"application/json", "application/xml"}
	chosenFormat, ok := contenttype.Negotiate(r.Header.Get("Accept"), supportedFormats)
	if !ok {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}

	// 2. Language Negotiation
	supportedLangs := []string{"en-US", "hi-IN"}
	chosenLang, _ := contenttype.NegotiateLanguage(r.Header.Get("Accept-Language"), supportedLangs)

	fmt.Printf("Serving response as %s in %s\n", chosenFormat, chosenLang)
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestParseContentType
--- PASS: TestParseContentType (0.00s)
=== RUN   TestParseAccept
--- PASS: TestParseAccept (0.00s)
=== RUN   TestNegotiate
--- PASS: TestNegotiate (0.00s)
=== RUN   TestNegotiateLanguage
--- PASS: TestNegotiateLanguage (0.00s)
=== RUN   TestNegotiateEncoding
--- PASS: TestNegotiateEncoding (0.00s)
=== RUN   TestErrors
--- PASS: TestErrors (0.00s)
PASS
ok  	go-contenttype	0.550s
```

---

## 📄 License
MIT License.
