package keyrotator_test

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-keyrotator"
)

func ExampleRotator_GetNextKey() {
	// Initialize rotator with 3 keys
	rotator := keyrotator.FromCSV("api-key-1,api-key-2,api-key-3")

	// 1. Fetch first key
	key1, _ := rotator.GetNextKey()
	fmt.Println("First key:", key1)

	// 2. Fetch second key (Round-Robin)
	key2, _ := rotator.GetNextKey()
	fmt.Println("Second key:", key2)

	// 3. Simulate rate-limiting on key-3
	rotator.MarkRateLimited("api-key-3", 10*time.Minute)

	// 4. Subsequent fetch skips rate-limited key and cycles back to key-1
	key3, _ := rotator.GetNextKey()
	fmt.Println("Third key:", key3)

	// Output:
	// First key: api-key-1
	// Second key: api-key-2
	// Third key: api-key-1
}
