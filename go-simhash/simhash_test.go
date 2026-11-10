package simhash_test

import (
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-simhash"
)

func TestSimHash_ExactAndNearDuplicate(t *testing.T) {
	doc1 := "The quick brown fox jumps over the lazy dog in the sunny afternoon"
	doc2 := "The quick brown fox jumps over the lazy dog in the sunny afternoon"
	doc3 := "The fast brown fox jumps over the lazy dog in the sunny afternoon" // minor change: quick -> fast
	unrelated := "PostgreSQL distributed consensus algorithm with raft and paxos replication across datacenters"

	fp1 := simhash.ComputeFromText(doc1)
	fp2 := simhash.ComputeFromText(doc2)
	fp3 := simhash.ComputeFromText(doc3)
	fpUnrelated := simhash.ComputeFromText(unrelated)

	// Exact match distance should be 0
	if dist := fp1.Distance(fp2); dist != 0 {
		t.Fatalf("expected exact distance 0, got %d", dist)
	}
	if sim := fp1.Similarity(fp2); sim != 1.0 {
		t.Fatalf("expected similarity 1.0, got %f", sim)
	}

	// Minor difference should have significantly smaller distance than unrelated
	distNear := fp1.Distance(fp3)
	distFar := fp1.Distance(fpUnrelated)

	if distNear >= distFar {
		t.Fatalf("expected near distance (%d) to be much lower than unrelated distance (%d)", distNear, distFar)
	}
	if distNear > 12 {
		t.Fatalf("expected small distance for near-duplicate, got %d", distNear)
	}
	if distFar < 20 {
		t.Fatalf("expected large distance for unrelated doc, got %d", distFar)
	}
}

func TestSimHash_IndexNearDuplicates(t *testing.T) {
	idx := simhash.NewIndex()

	idx.Insert("article-1", simhash.ComputeFromText("Distributed systems and fault-tolerant replication"))
	idx.Insert("article-2", simhash.ComputeFromText("Distributed systems and fault tolerant replication")) // tiny typo
	idx.Insert("article-3", simhash.ComputeFromText("Recipe for homemade Italian pizza with mozzarella"))

	if idx.Count() != 3 {
		t.Fatalf("expected 3 items, got %d", idx.Count())
	}

	query := simhash.ComputeFromText("Distributed systems and fault-tolerant replication")
	matches := idx.FindNearDuplicates(query, 5)

	if len(matches) != 2 {
		t.Fatalf("expected 2 near-duplicate matches, got %d", len(matches))
	}

	// Article 3 should not match
	for _, m := range matches {
		if m.ID == "article-3" {
			t.Fatal("unrelated article should not match")
		}
	}
}

func TestSimHash_Concurrency(t *testing.T) {
	idx := simhash.NewIndex()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			fp := simhash.ComputeFromText("Sample content for worker concurrency test")
			idx.Insert("worker", fp)
			_ = idx.FindNearDuplicates(fp, 3)
		}(i)
	}

	wg.Wait()
}

