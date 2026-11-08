# Technical Specification: go-jsonstream

## 1. Domain & Purpose
`go-jsonstream` is a high-performance, OOM-proof streaming parser and writer for massive JSON arrays and newline-delimited JSON (NDJSON) streams in Go. It enables processing multi-gigabyte datasets with constant, near-zero memory footprint by tokenizing and decoding items incrementally from any `io.Reader`.

## 2. Architecture & Design Principles
- **Token-Driven Incremental Decoding**: Leverages standard `encoding/json.Decoder` token stream (`[` ... `]`) to decode elements without holding full payloads in RAM.
- **Type-Safe Generics**: Decode and encode arbitrary types `T` with full compile-time safety.
- **Dual Consumption Patterns**:
  - Callback-based `StreamArray(r, fn)` for early-exit fail-fast workflows.
  - Channel-based `StreamArrayChan(ctx, r, bufSize)` for asynchronous pipeline processing with context cancellation.
- **NDJSON Support**: Streaming line-by-line decoding and encoding for log ingestion and data lake pipelines.
- **Low-Allocation Streaming Writers**: `ArrayWriter[T]` and `NDJSONWriter[T]` allow writing continuous streaming responses directly to `http.ResponseWriter` or disk.
- **Zero Heavy Dependencies**: Pure standard library (`encoding/json`, `bufio`, `io`, `sync`, `context`).

## 3. API Surface
- `type Result[T any] struct { Value T; Err error }`
- `func StreamArray[T any](r io.Reader, fn func(item T) error) error`
- `func StreamArrayChan[T any](ctx context.Context, r io.Reader, bufSize int) <-chan Result[T]`
- `func StreamNDJSON[T any](r io.Reader, fn func(item T) error) error`
- `func StreamNDJSONChan[T any](ctx context.Context, r io.Reader, bufSize int) <-chan Result[T]`
- `type ArrayWriter[T any] struct`
- `func NewArrayWriter[T any](w io.Writer) *ArrayWriter[T]`
- `func (aw *ArrayWriter[T]) Write(item T) error`
- `func (aw *ArrayWriter[T]) Close() error`
- `type NDJSONWriter[T any] struct`
- `func NewNDJSONWriter[T any](w io.Writer) *NDJSONWriter[T]`
- `func (nw *NDJSONWriter[T]) Write(item T) error`

## 4. Error Handling
- `ErrExpectedArray`: Input does not begin with an array opening bracket `[`.
- `ErrMalformedJSON`: Input contains invalid syntax or ends prematurely.
- `ErrWriterClosed`: Writing attempted to an already finalized `ArrayWriter`.

## 5. Verification & Testing
- Unit tests verify callback array parsing, channel-based streaming, NDJSON decoding, array writer roundtrips, and syntax error detections.
