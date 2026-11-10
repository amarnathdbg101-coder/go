package merkle

import (
	"encoding/hex"
)

// Tree represents a cryptographically verifiable RFC 6962 Binary Merkle Tree.
type Tree struct {
	leaves [][]byte
	levels [][][]byte
}

// NewTree constructs a Merkle Tree from raw byte slices.
func NewTree(data [][]byte) (*Tree, error) {
	if len(data) == 0 {
		return nil, ErrEmptyTree
	}

	leaves := make([][]byte, len(data))
	for i, d := range data {
		leaves[i] = HashLeaf(d)
	}

	levels := [][][]byte{leaves}
	currLevel := leaves

	for len(currLevel) > 1 {
		var nextLevel [][]byte
		for i := 0; i < len(currLevel); i += 2 {
			if i+1 < len(currLevel) {
				nextLevel = append(nextLevel, HashNode(currLevel[i], currLevel[i+1]))
			} else {
				// Odd leaf handling: duplicate last hash
				nextLevel = append(nextLevel, HashNode(currLevel[i], currLevel[i]))
			}
		}
		levels = append(levels, nextLevel)
		currLevel = nextLevel
	}

	return &Tree{
		leaves: data,
		levels: levels,
	}, nil
}

// Root returns the 32-byte cryptographic Merkle Root Hash.
func (t *Tree) Root() []byte {
	top := t.levels[len(t.levels)-1]
	return top[0]
}

// RootHex returns the hexadecimal string representation of the root hash.
func (t *Tree) RootHex() string {
	return hex.EncodeToString(t.Root())
}

// LeafCount returns the number of leaves in the tree.
func (t *Tree) LeafCount() int {
	return len(t.leaves)
}

// GenerateProof creates an O(log N) audit path proof for the leaf at index.
func (t *Tree) GenerateProof(index int) (*Proof, error) {
	if index < 0 || index >= len(t.leaves) {
		return nil, ErrIndexOutOfBounds
	}

	var path []ProofNode
	currIdx := index

	for l := 0; l < len(t.levels)-1; l++ {
		level := t.levels[l]
		isRightSibling := (currIdx%2 == 0)

		var siblingIdx int
		if isRightSibling {
			siblingIdx = currIdx + 1
		} else {
			siblingIdx = currIdx - 1
		}

		var siblingHash []byte
		if siblingIdx < len(level) {
			siblingHash = level[siblingIdx]
		} else {
			// Sibling was duplicated self
			siblingHash = level[currIdx]
		}

		path = append(path, ProofNode{
			Hash:    siblingHash,
			IsRight: isRightSibling,
		})

		currIdx /= 2
	}

	return &Proof{
		LeafIndex: index,
		AuditPath: path,
	}, nil
}
