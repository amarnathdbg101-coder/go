# go-jsonstream 🌊

`go-jsonstream` ek high-performance, memory-efficient streaming JSON aur NDJSON (Newline Delimited JSON) parser & writer library hai Go ke liye. Ye multi-gigabyte files aur continuous streams ko bina OOM (Out Of Memory) crash ke item-by-item decode aur encode karta hai directly over `io.Reader` and `io.Writer`.

---

## ✨ Features

- **OOM-Proof Streaming**: Infinite size ke JSON arrays ya NDJSON logs ko constant, near-zero RAM usage ke saath process karein.
- **Type-Safe Generics**: Strongly typed struct decoding (`StreamArray[Product]()`).
- **Dual Processing Patterns**:
  - Direct Callback (`StreamArray`) for inline processing.
  - Context-Aware Channel (`StreamArrayChan`) with buffered queues & cancellation.
- **NDJSON Stream Support**: High-volume log pipelines aur Kafka/Elasticsearch export ingestion.
- **Streaming Writers**: Direct HTTP chunked streaming ya file writing via `ArrayWriter` and `NDJSONWriter`.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-jsonstream
```

### Stream 1GB JSON Array from File or Network
```go
package main

import (
	"fmt"
	"os"

	"github.com/shopsilo/go-jsonstream"
)

type Customer struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func main() {
	file, err := os.Open("massive_customers_export.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Decodes customers one-by-one without loading the whole file into RAM
	count := 0
	err = jsonstream.StreamArray(file, func(c Customer) error {
		count++
		if count%10000 == 0 {
			fmt.Printf("Processed %d customers...\n", count)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Stream terminated: %v\n", err)
	}
	fmt.Printf("Successfully imported %d customers with constant RAM!\n", count)
}
```

### Stream HTTP Response
```go
func streamProductsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    writer := jsonstream.NewArrayWriter[Product](w)
    defer writer.Close()

    for product := range fetchProductsFromDB() {
        if err := writer.Write(product); err != nil {
            return
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
=== RUN   TestStreamArray
--- PASS: TestStreamArray (0.00s)
=== RUN   TestStreamArrayChan
--- PASS: TestStreamArrayChan (0.00s)
=== RUN   TestStreamNDJSON
--- PASS: TestStreamNDJSON (0.00s)
=== RUN   TestArrayWriter
--- PASS: TestArrayWriter (0.00s)
=== RUN   TestNDJSONWriter
--- PASS: TestNDJSONWriter (0.00s)
=== RUN   TestErrors
--- PASS: TestErrors (0.00s)
PASS
ok  	go-jsonstream	0.559s
```

---

## 📄 License
MIT License.
