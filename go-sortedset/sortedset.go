package sortedset

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
)

const (
	maxLevel    = 32
	probability = 0.25
)

// Item represents a member and its associated numeric score.
type Item[M comparable] struct {
	Member M       `json:"member"`
	Score  float64 `json:"score"`
}

type skipListNode[M comparable] struct {
	item     Item[M]
	backward *skipListNode[M]
	level    []skipListLevel[M]
}

type skipListLevel[M comparable] struct {
	forward *skipListNode[M]
	span    int
}

type skipList[M comparable] struct {
	header *skipListNode[M]
	tail   *skipListNode[M]
	length int
	level  int
}

func newSkipList[M comparable]() *skipList[M] {
	sl := &skipList[M]{
		level: 1,
	}
	sl.header = &skipListNode[M]{
		level: make([]skipListLevel[M], maxLevel),
	}
	return sl
}

func randomLevel() int {
	lvl := 1
	var b [8]byte
	_, _ = rand.Read(b[:])
	r := binary.LittleEndian.Uint64(b[:])

	for (r&0xFFFF) < (0xFFFF/4) && lvl < maxLevel {
		lvl++
		r >>= 16
	}
	return lvl
}

func (sl *skipList[M]) insert(item Item[M]) *skipListNode[M] {
	var update [maxLevel]*skipListNode[M]
	var rank [maxLevel]int

	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.level[i].forward != nil &&
			(x.level[i].forward.item.Score < item.Score ||
				(x.level[i].forward.item.Score == item.Score && x.level[i].forward.item.Member != item.Member)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	lvl := randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].level[i].span = sl.length
		}
		sl.level = lvl
	}

	x = &skipListNode[M]{
		item:  item,
		level: make([]skipListLevel[M], lvl),
	}

	for i := 0; i < lvl; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	for i := lvl; i < sl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == sl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		sl.tail = x
	}

	sl.length++
	return x
}

func (sl *skipList[M]) deleteNode(x *skipListNode[M], update [maxLevel]*skipListNode[M]) {
	for i := 0; i < sl.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		sl.tail = x.backward
	}

	for sl.level > 1 && sl.header.level[sl.level-1].forward == nil {
		sl.level--
	}
	sl.length--
}

func (sl *skipList[M]) delete(item Item[M]) bool {
	var update [maxLevel]*skipListNode[M]
	x := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.item.Score < item.Score ||
				(x.level[i].forward.item.Score == item.Score && x.level[i].forward.item.Member != item.Member)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward
	if x != nil && x.item.Score == item.Score && x.item.Member == item.Member {
		sl.deleteNode(x, update)
		return true
	}
	return false
}

func (sl *skipList[M]) getRank(item Item[M]) int {
	rank := 0
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.item.Score < item.Score ||
				(x.level[i].forward.item.Score == item.Score && x.level[i].forward.item.Member != item.Member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
		if x.level[i].forward != nil && x.level[i].forward.item.Member == item.Member {
			return rank + x.level[i].span
		}
	}
	return 0
}

func (sl *skipList[M]) getElementByRank(rank int) *skipListNode[M] {
	traversed := 0
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

// Set represents a fast Redis-style sorted set of unique members indexed by numeric score.
type Set[M comparable] struct {
	mu         sync.RWMutex
	concurrent bool
	sl         *skipList[M]
	dict       map[M]float64
}

// New creates a new non-concurrent SortedSet.
func New[M comparable]() *Set[M] {
	return &Set[M]{
		sl:   newSkipList[M](),
		dict: make(map[M]float64),
	}
}

// NewConcurrent creates a thread-safe SortedSet guarded by sync.RWMutex.
func NewConcurrent[M comparable]() *Set[M] {
	return &Set[M]{
		concurrent: true,
		sl:         newSkipList[M](),
		dict:       make(map[M]float64),
	}
}

// Add inserts or updates a member's score in O(log N).
// Returns true if a new member was added, false if an existing member's score was updated.
func (s *Set[M]) Add(member M, score float64) bool {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	oldScore, exists := s.dict[member]
	if exists {
		if oldScore == score {
			return false
		}
		// Remove old position in skip list
		s.sl.delete(Item[M]{Member: member, Score: oldScore})
	}

	s.dict[member] = score
	s.sl.insert(Item[M]{Member: member, Score: score})
	return !exists
}

// Score retrieves a member's score in O(1) time.
func (s *Set[M]) Score(member M) (float64, bool) {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}

	score, ok := s.dict[member]
	return score, ok
}

// Rank returns the 0-based rank from lowest to highest score in O(log N).
func (s *Set[M]) Rank(member M) (int, bool) {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}

	score, exists := s.dict[member]
	if !exists {
		return -1, false
	}

	rank := s.sl.getRank(Item[M]{Member: member, Score: score})
	if rank == 0 {
		return -1, false
	}
	return rank - 1, true
}

// RevRank returns the 0-based rank from highest to lowest score in O(log N) (ideal for leaderboards).
func (s *Set[M]) RevRank(member M) (int, bool) {
	rank, exists := s.Rank(member)
	if !exists {
		return -1, false
	}
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	return s.sl.length - 1 - rank, true
}

// Delete removes a member from the sorted set in O(log N).
func (s *Set[M]) Delete(member M) bool {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	score, exists := s.dict[member]
	if !exists {
		return false
	}

	delete(s.dict, member)
	return s.sl.delete(Item[M]{Member: member, Score: score})
}

// Len returns the total number of members in the sorted set.
func (s *Set[M]) Len() int {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	return len(s.dict)
}

// RangeByRank returns members between 0-based start and stop ranks inclusive.
func (s *Set[M]) RangeByRank(start, stop int) []Item[M] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}

	if s.sl.length == 0 {
		return nil
	}

	if start < 0 {
		start = s.sl.length + start
	}
	if stop < 0 {
		stop = s.sl.length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= s.sl.length {
		stop = s.sl.length - 1
	}

	if start > stop {
		return nil
	}

	// 1-based rank for skip list query
	node := s.sl.getElementByRank(start + 1)
	count := stop - start + 1
	result := make([]Item[M], 0, count)

	for node != nil && count > 0 {
		result = append(result, node.item)
		node = node.level[0].forward
		count--
	}

	return result
}

// RangeByScore returns members whose score is between min and max inclusive.
func (s *Set[M]) RangeByScore(minScore, maxScore float64) []Item[M] {
	if s.concurrent {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}

	var result []Item[M]
	curr := s.sl.header.level[0].forward

	// Advance to first element >= minScore
	for curr != nil && curr.item.Score < minScore {
		curr = curr.level[0].forward
	}

	for curr != nil && curr.item.Score <= maxScore {
		result = append(result, curr.item)
		curr = curr.level[0].forward
	}

	return result
}

// PopMin removes and returns the member with lowest score.
func (s *Set[M]) PopMin() (*Item[M], bool) {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	if s.sl.length == 0 {
		return nil, false
	}

	first := s.sl.header.level[0].forward
	item := first.item

	delete(s.dict, item.Member)
	s.sl.delete(item)
	return &item, true
}

// PopMax removes and returns the member with highest score.
func (s *Set[M]) PopMax() (*Item[M], bool) {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	if s.sl.length == 0 {
		return nil, false
	}

	tail := s.sl.tail
	item := tail.item

	delete(s.dict, item.Member)
	s.sl.delete(item)
	return &item, true
}

// Clear removes all elements from the sorted set.
func (s *Set[M]) Clear() {
	if s.concurrent {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.sl = newSkipList[M]()
	s.dict = make(map[M]float64)
}
