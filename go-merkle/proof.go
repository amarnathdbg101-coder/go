package merkle

import (
	"bytes"
	"crypto/sha256"
)

// ProofNode represents a sibling hash step along an audit path.
type ProofNode struct {
	Hash    []byte `json:"hash"`
	IsRight bool   `json:"is_right"`
}

// Proof contains the logarithmic audit path necessary to prove leaf membership.
type Proof struct {
	LeafIndex int         `json:"leaf_index"`
	AuditPath []ProofNode `json:"audit_path"`
}

// Verify mathematically validates whether leaf is part of the Merkle Tree with the given rootHash.
func (p *Proof) Verify(leaf []byte, rootHash []byte) bool {
	if len(rootHash) != 32 {
		return false
	}

	curr := HashLeaf(leaf)

	for _, step := range p.AuditPath {
		if step.IsRight {
			curr = HashNode(curr, step.Hash)
		} else {
			curr = HashNode(step.Hash, curr)
		}
	}

	return bytes.Equal(curr, rootHash)
}

// HashLeaf computes RFC 6962 leaf hash: SHA-256(0x00 || data).
func HashLeaf(data []byte) []byte {
	h := sha256.New()
	h.Write([]byte{0x00})
	h.Write(data)
	return h.Sum(nil)
}

// HashNode computes RFC 6962 interior node hash: SHA-256(0x01 || left || right).
func HashNode(left, right []byte) []byte {
	h := sha256.New()
	h.Write([]byte{0x01})
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}
