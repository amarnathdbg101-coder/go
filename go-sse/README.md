# go-sse 📡

`go-sse` ek lightweight, WHATWG EventSource compliant Server-Sent Events (SSE) streaming engine hai Go ke liye. Ye HTTP connection par real-time live events, AI token streams (LLM streaming), order updates, aur dashboard push notifications ke liye banaya gaya hai without WebSocket complexity.

---

## ✨ Features

- **Full WHATWG SSE Standard**: Supports `id:`, `event:`, `data:`, `retry:`, aur keep-alive `: ping` comments.
- **Auto Multi-Line & JSON Encoding**: Multi-line strings ko automatically `data:` me split karta hai aur Go structs ko directly JSON me marshal karta hai.
- **Nginx Reverse-Proxy Safe**: Injects `X-Accel-Buffering: no` taaki intermediaries buffering na karein.
- **Client Disconnect Detection**: `s.Done()` and context cancellation se resources immediately release hote hain.
- **Thread-Safe Broadcasting**: Multiple goroutines safely event stream par write kar sakti hain.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-sse
```

### Complete Live Event Streaming Handler
```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shopsilo/go-sse"
)

type StockUpdate struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

func stockStreamHandler(w http.ResponseWriter, r *http.Request) {
	streamer, err := sse.NewStreamer(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-streamer.Done():
			fmt.Println("Client disconnected!")
			return

		case t := <-ticker.C:
			// Send typed event with ID and struct payload
			err := streamer.Send(sse.Event{
				ID:    fmt.Sprintf("%d", t.Unix()),
				Event: "price_update",
				Data: StockUpdate{
					Symbol: "RELIANCE",
					Price:  2940.50,
				},
				Retry: 3 * time.Second,
			})
			if err != nil {
				return
			}
		}
	}
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestNewStreamer_Headers
--- PASS: TestNewStreamer_Headers (0.00s)
=== RUN   TestNewStreamer_NonFlusher
--- PASS: TestNewStreamer_NonFlusher (0.00s)
=== RUN   TestStreamer_SendEvent
--- PASS: TestStreamer_SendEvent (0.00s)
=== RUN   TestStreamer_JSONData
--- PASS: TestStreamer_JSONData (0.00s)
=== RUN   TestStreamer_Ping
--- PASS: TestStreamer_Ping (0.00s)
=== RUN   TestStreamer_ContextCancellation
--- PASS: TestStreamer_ContextCancellation (0.00s)
PASS
ok  	go-sse	0.729s
```

---

## 📄 License
MIT License.
