# go-hpack

`go-hpack` provides a production-grade, RFC 7541 compliant HPACK header compression and decompression engine in pure Go. It is the core networking primitive behind HTTP/2 and gRPC transport protocols, compressing bulky HTTP request and response headers from hundreds of bytes down to single-digit bytes.

## Key Features

- **RFC 7541 Compliance**: Full implementation of Static Table (61 entries), Dynamic Table with FIFO eviction, and Integer Representation encoding.
- **Extreme Compression**: Common headers like `:method: GET` or `:status: 200` compress down to exactly **1 single byte** (`0x82`, `0x88`).
- **Dynamic Table Cache**: Automatically learns and indexes repetitive headers throughout an HTTP/2 connection session.
- **Sensitive Header Protection**: Native support for "Never Indexed" literal headers (`Sensitive: true`) preventing tokens and passwords from entering shared decompression buffers (defense against CRIME/BREACH style side-channel attacks).
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-hpack"
)

func main() {
	enc := hpack.NewEncoder(4096)
	dec := hpack.NewDecoder(4096)

	headers := []hpack.HeaderField{
		{Name: ":method", Value: "GET"},
		{Name: ":path", Value: "/"},
		{Name: ":scheme", Value: "https"},
		{Name: ":status", Value: "200"},
	}

	// Compress
	compressed := enc.Encode(headers)
	fmt.Printf("Original: 4 headers | Compressed: %d bytes (1 byte per header!)\n", len(compressed))

	// Decompress
	decoded, _ := dec.Decode(compressed)
	for _, h := range decoded {
		fmt.Printf("  %s: %s\n", h.Name, h.Value)
	}
}
```
