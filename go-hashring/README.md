# go-hashring 💍

`go-hashring` ek high-performance, thread-safe consistent hash ring library hai Go ke liye jo virtual nodes (Ketama algorithm) use karti hai. Ye distributed caching (Redis/Memcached clusters), database sharding, aur load balancing ke liye banayi gayi hai taaki nodes add ya remove hone par minimal key migration ($K/N$) ho aur cache stampedes prevent hon.

---

## ✨ Features

- **Minimal Key Migration**: Node crash ya scaling ke waqt pure cluster ke keys rehash hone ke bajaye sirf $1/N$ keys move hoti hain.
- **Uniform Distribution with Virtual Nodes**: 64-bit SplitMix64 avalanche bit mixer hotspot skew ko eliminate karta hai.
- **Multi-Replica Routing (`GetN`)**: Clockwise order me $N$ distinct replica nodes nikalta hai for distributed replication.
- **Dynamic Cluster Membership**: Live traffic ke dauraan `Add` aur `Remove` nodes safely call karein (`sync.RWMutex`).
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-hashring
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-hashring"
)

func main() {
	// Initialize ring with 100 virtual nodes per server
	ring := hashring.New(100, "redis-node-1", "redis-node-2", "redis-node-3")

	// Route keys deterministically
	userKey := "user_session:10421"
	targetNode, err := ring.Get(userKey)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Key %s routed to -> %s\n", userKey, targetNode)

	// Get 2 replica nodes for high availability
	replicas, _ := ring.GetN(userKey, 2)
	fmt.Printf("Primary: %s, Secondary: %s\n", replicas[0], replicas[1])

	// Dynamic cluster scale-out
	ring.Add("redis-node-4")
	fmt.Printf("Cluster expanded. Total nodes: %d\n", ring.Len())
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestHashRing_BasicGetAndDeterminism
--- PASS: TestHashRing_BasicGetAndDeterminism (0.00s)
=== RUN   TestHashRing_MinimalKeyMigration
--- PASS: TestHashRing_MinimalKeyMigration (0.00s)
=== RUN   TestHashRing_GetNReplicas
--- PASS: TestHashRing_GetNReplicas (0.00s)
=== RUN   TestHashRing_Empty
--- PASS: TestHashRing_Empty (0.00s)
=== RUN   TestHashRing_Concurrency
--- PASS: TestHashRing_Concurrency (0.00s)
PASS
ok  	go-hashring	0.478s
```

---

## 📄 License
MIT License.
