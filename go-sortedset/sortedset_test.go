package sortedset

import (
	"sync"
	"testing"
)

func TestSortedSet_Basic(t *testing.T) {
	ss := New[string]()

	// Add players
	ss.Add("player1", 100)
	ss.Add("player2", 250)
	ss.Add("player3", 50)
	ss.Add("player4", 400)

	if ss.Len() != 4 {
		t.Fatalf("expected len 4, got %d", ss.Len())
	}

	score, ok := ss.Score("player2")
	if !ok || score != 250 {
		t.Fatalf("expected score 250, got %f, ok: %v", score, ok)
	}

	// Lowest score rank: player3 (50) -> rank 0, player1 (100) -> rank 1, player2 (250) -> rank 2, player4 (400) -> rank 3
	rank, ok := ss.Rank("player3")
	if !ok || rank != 0 {
		t.Fatalf("expected rank 0 for player3, got %d", rank)
	}
	rank, ok = ss.Rank("player4")
	if !ok || rank != 3 {
		t.Fatalf("expected rank 3 for player4, got %d", rank)
	}

	// Leaderboard RevRank: player4 (400) -> revRank 0, player2 (250) -> revRank 1, player1 (100) -> revRank 2, player3 (50) -> revRank 3
	revRank, ok := ss.RevRank("player4")
	if !ok || revRank != 0 {
		t.Fatalf("expected revRank 0 for top player, got %d", revRank)
	}
	revRank, ok = ss.RevRank("player3")
	if !ok || revRank != 3 {
		t.Fatalf("expected revRank 3 for lowest player, got %d", revRank)
	}
}

func TestSortedSet_UpdateScore(t *testing.T) {
	ss := New[string]()
	ss.Add("alice", 100)
	ss.Add("bob", 200)

	// Update alice's score to 300 -> becomes #1
	updated := ss.Add("alice", 300)
	if updated {
		t.Fatalf("expected Add to return false on update")
	}

	revRank, _ := ss.RevRank("alice")
	if revRank != 0 {
		t.Fatalf("expected alice to be #1 after score update, got %d", revRank)
	}
}

func TestSortedSet_RangeQueries(t *testing.T) {
	ss := New[string]()
	ss.Add("a", 10)
	ss.Add("b", 20)
	ss.Add("c", 30)
	ss.Add("d", 40)
	ss.Add("e", 50)

	// RangeByScore [20, 40] -> b, c, d
	byScore := ss.RangeByScore(20, 40)
	if len(byScore) != 3 {
		t.Fatalf("expected 3 items in score range [20, 40], got %d", len(byScore))
	}
	if byScore[0].Member != "b" || byScore[2].Member != "d" {
		t.Fatalf("unexpected range by score items: %+v", byScore)
	}

	// RangeByRank [1, 3] -> rank 1 (b), rank 2 (c), rank 3 (d)
	byRank := ss.RangeByRank(1, 3)
	if len(byRank) != 3 {
		t.Fatalf("expected 3 items in rank range [1, 3], got %d", len(byRank))
	}
	if byRank[0].Member != "b" || byRank[2].Member != "d" {
		t.Fatalf("unexpected range by rank items: %+v", byRank)
	}
}

func TestSortedSet_PopMinMax(t *testing.T) {
	ss := New[int]()
	ss.Add(1, 100)
	ss.Add(2, 500)
	ss.Add(3, 50)

	minItem, ok := ss.PopMin()
	if !ok || minItem.Member != 3 || minItem.Score != 50 {
		t.Fatalf("unexpected PopMin result: %+v", minItem)
	}

	maxItem, ok := ss.PopMax()
	if !ok || maxItem.Member != 2 || maxItem.Score != 500 {
		t.Fatalf("unexpected PopMax result: %+v", maxItem)
	}

	if ss.Len() != 1 {
		t.Fatalf("expected len 1 after pops, got %d", ss.Len())
	}
}

func TestSortedSet_Delete(t *testing.T) {
	ss := New[string]()
	ss.Add("x", 10)
	ss.Add("y", 20)

	if !ss.Delete("x") {
		t.Fatalf("expected Delete to return true")
	}
	if ss.Delete("x") {
		t.Fatalf("expected second Delete to return false")
	}
	if ss.Len() != 1 {
		t.Fatalf("expected len 1, got %d", ss.Len())
	}
}

func TestSortedSet_Concurrent(t *testing.T) {
	ss := NewConcurrent[int]()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ss.Add(id, float64(id*10))
			ss.Score(id)
			ss.Rank(id)
			ss.RevRank(id)
		}(i)
	}
	wg.Wait()

	if ss.Len() != 50 {
		t.Fatalf("expected len 50, got %d", ss.Len())
	}
}
