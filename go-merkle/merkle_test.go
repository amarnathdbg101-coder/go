package merkle_test

import (
	"bytes"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-merkle"
)

func TestMerkleTreeDeterministic(t *testing.T) {
	data := [][]byte{
		[]byte("tx_1: Amardeep paid ₹500 to ShopSilo"),
		[]byte("tx_2: Supplier invoice settled ₹12000"),
		[]byte("tx_3: Cashier closed drawer ₹4500"),
	}

	tree1, err := merkle.NewTree(data)
	if err != nil {
		t.Fatalf("failed to create tree1: %v", err)
	}

	tree2, err := merkle.NewTree(data)
	if err != nil {
		t.Fatalf("failed to create tree2: %v", err)
	}

	if !bytes.Equal(tree1.Root(), tree2.Root()) {
		t.Fatalf("merkle root must be strictly deterministic")
	}
}

func TestAuditProofVerification(t *testing.T) {
	data := [][]byte{
		[]byte("entry_0"),
		[]byte("entry_1"),
		[]byte("entry_2"),
		[]byte("entry_3"),
		[]byte("entry_4"),
	}

	tree, err := merkle.NewTree(data)
	if err != nil {
		t.Fatalf("tree creation failed: %v", err)
	}

	rootHash := tree.Root()

	// Generate audit proof for leaf #2 ("entry_2")
	proof, err := tree.GenerateProof(2)
	if err != nil {
		t.Fatalf("failed to generate proof: %v", err)
	}

	// 1. Verify valid leaf
	if !proof.Verify([]byte("entry_2"), rootHash) {
		t.Fatalf("proof verification should succeed for valid entry")
	}

	// 2. Tampered leaf must fail verification!
	if proof.Verify([]byte("entry_2_TAMPERED"), rootHash) {
		t.Fatalf("proof verification MUST fail for tampered data")
	}
}

