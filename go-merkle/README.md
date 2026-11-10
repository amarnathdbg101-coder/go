# go-merkle 🌳

> **RFC 6962-Compliant Cryptographic Binary Merkle Tree & Audit Proof Engine for Go (Go 1.23+).**  
> Generate tamper-proof ledger root hashes and verify cryptographic membership proofs in $O(\log N)$ time without sharing full datasets.

---

## 🚀 Key Features

1. **RFC 6962 Domain Separation**:
   - Leaf hash: $\text{SHA-256}(0x00 \ || \ \text{data})$
   - Node hash: $\text{SHA-256}(0x01 \ || \ \text{left} \ || \ \text{right})$
   - Mathematically shields against length-extension and second-preimage collision attacks.
2. **Logarithmic Proofs ($O(\log N)$)**:
   - Prove that a single transaction exists in a million-record block with an audit path of just $\sim 20$ hashes.
3. **Odd-Leaf Resilient**:
   - Adheres to Bitcoin / Certificate Transparency RFC rules for odd leaves at each tree height.
4. **Zero Dependencies**:
   - 100% pure standard library (`crypto/sha256`, `bytes`, `encoding/hex`).

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-merkle"
)

func main() {
	// Transactions batch
	txs := [][]byte{
		[]byte("TX-101: ₹500 to Merchant"),
		[]byte("TX-102: ₹1200 to Supplier"),
		[]byte("TX-103: ₹250 Refund"),
	}

	tree, _ := merkle.NewTree(txs)
	fmt.Println("Merkle Root:", tree.RootHex())

	// Generate inclusion proof for TX-102 (index 1)
	proof, _ := tree.GenerateProof(1)

	// Verify membership against RootHash
	isValid := proof.Verify([]byte("TX-102: ₹1200 to Supplier"), tree.Root())
	fmt.Println("Cryptographic Proof Verified:", isValid) // true
}
```
