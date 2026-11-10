# SPEC: go-merkle

## 1. Scope & Purpose
`go-merkle` provides cryptographic verification of data integrity and dataset inclusion proofs using Binary Merkle Trees. It serves audit logs, transactional ledgers, and zero-knowledge data verification.

It does NOT provide decentralized consensus or blockchain mining mechanisms.

## 2. Public API Contract
```go
type ProofNode struct { Hash []byte, IsRight bool }
type Proof struct { LeafIndex int, AuditPath []ProofNode }
type Tree struct { ... }

func NewTree(data [][]byte) (*Tree, error)
func (t *Tree) Root() []byte
func (t *Tree) RootHex() string
func (t *Tree) LeafCount() int
func (t *Tree) GenerateProof(index int) (*Proof, error)
func (p *Proof) Verify(leaf []byte, rootHash []byte) bool
```

## 3. Concurrency & Memory Guarantees
- Tree structures are immutable once generated.
- Proof verification is lock-free and runs in $O(\log N)$ hash steps.
- Domain separation prefixes prevent leaf/internal node hash confusion.

## 4. How to Extend
- Implement Sparse Merkle Trees (SMT) for verifiable key-value state proofs.
- Add persistent storage adapters for multi-gigabyte append-only trees.
