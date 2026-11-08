# Technical Specification: go-sse

## 1. Domain & Purpose
`go-sse` is a WHATWG EventSource compliant Server-Sent Events (SSE) streaming engine for Go. It enables bidirectional-friendly, real-time push streaming over standard HTTP connections for live notifications, AI token completion streams, financial tickers, and dashboard updates.

## 2. Architecture & Design Principles
- **WHATWG Protocol Compliance**: Supports all standard fields: `id`, `event`, `data`, `retry` (milliseconds), and keep-alive comments (`: ping`).
- **Automatic Multi-Line & JSON Formatting**: Seamlessly chunks multiline strings into repeated `data:` lines and auto-serializes complex Go structs to JSON.
- **Nginx & Reverse-Proxy Safety**: Automatically sets `X-Accel-Buffering: no`, `Cache-Control: no-cache`, and `Content-Type: text/event-stream`.
- **Disconnect & Cancellation Detection**: Listens directly to `r.Context().Done()` to prevent thread leaks when browsers disconnect.
- **Thread Safety**: Synchronized `Send()` calls allow multiple worker goroutines to push events over the same connection.
- **Zero Heavy Dependencies**: Pure standard library (`net/http`, `encoding/json`, `bytes`, `strings`, `sync`, `time`).

## 3. API Surface
- `type Event struct`
- `type Streamer struct`
- `func NewStreamer(w http.ResponseWriter, r *http.Request) (*Streamer, error)`
- `func (s *Streamer) Send(e Event) error`
- `func (s *Streamer) SendData(data any) error`
- `func (s *Streamer) Ping() error`
- `func (s *Streamer) Close()`
- `func (s *Streamer) Done() <-chan struct{}`

## 4. Error Handling
- `ErrStreamingNotSupported`: Returned when `http.ResponseWriter` does not implement `http.Flusher`.
- `ErrStreamClosed`: Returned when sending on a closed or disconnected stream.

## 5. Verification & Testing
- Unit tests verify header propagation, multi-line chunking, struct JSON marshaling, keep-alive comments, client disconnection detection, and flusher type assertions.
