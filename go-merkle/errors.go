package merkle

import "errors"

var (
	// ErrEmptyTree is returned when attempting to construct a Merkle Tree with zero leaves.
	ErrEmptyTree = errors.New("merkle: cannot construct a tree with empty leaf slice")

	// ErrIndexOutOfBounds is returned when requesting an audit proof for an invalid leaf index.
	ErrIndexOutOfBounds = errors.New("merkle: leaf index out of bounds")

	// ErrInvalidProof is returned when a Merkle proof fails mathematical verification against the root hash.
	ErrInvalidProof = errors.New("merkle: proof verification failed (hash mismatch)")
)
