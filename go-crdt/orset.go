package crdt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ORSet implements an Observed-Remove Set (Add-Wins Set) CRDT.
// It allows elements to be added and removed multiple times across partitioned nodes
// without ever encountering tombstone resurrection anomalies.
type ORSet[T comparable] struct {
	mu        sync.RWMutex
	nodeID    string
	addSet    map[T]map[string]struct{}
	removeSet map[T]map[string]struct{}
}

// NewORSet initializes an Observed-Remove Set for the specified node ID.
func NewORSet[T comparable](nodeID string) *ORSet[T] {
	if nodeID == "" {
		nodeID = "node_default"
	}
	return &ORSet[T]{
		nodeID:    nodeID,
		addSet:    make(map[T]map[string]struct{}),
		removeSet: make(map[T]map[string]struct{}),
	}
}

// Add inserts an element with a unique observation tag.
func (s *ORSet[T]) Add(elem T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag := s.generateTag()
	if _, exists := s.addSet[elem]; !exists {
		s.addSet[elem] = make(map[string]struct{})
	}
	s.addSet[elem][tag] = struct{}{}
}

// Remove removes an element by recording all currently observed tags into the remove set.
func (s *ORSet[T]) Remove(elem T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tags, exists := s.addSet[elem]
	if !exists || len(tags) == 0 {
		return
	}

	if _, ok := s.removeSet[elem]; !ok {
		s.removeSet[elem] = make(map[string]struct{})
	}

	for tag := range tags {
		s.removeSet[elem][tag] = struct{}{}
	}
}

// Contains checks if the element is currently a member of the set.
// It returns true if there is at least one tag in addSet that is not in removeSet.
func (s *ORSet[T]) Contains(elem T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addTags, exists := s.addSet[elem]
	if !exists || len(addTags) == 0 {
		return false
	}

	remTags := s.removeSet[elem]
	for tag := range addTags {
		if _, removed := remTags[tag]; !removed {
			return true
		}
	}
	return false
}

// Elements returns a slice of all active elements in the set.
func (s *ORSet[T]) Elements() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []T
	for elem, addTags := range s.addSet {
		remTags := s.removeSet[elem]
		hasActive := false
		for tag := range addTags {
			if _, removed := remTags[tag]; !removed {
				hasActive = true
				break
			}
		}
		if hasActive {
			result = append(result, elem)
		}
	}
	return result
}

// Merge converges this ORSet with another ORSet by taking the union of addSets and removeSets.
func (s *ORSet[T]) Merge(other *ORSet[T]) error {
	if other == nil {
		return ErrNilNode
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	// Union addSets
	for elem, otherTags := range other.addSet {
		if _, exists := s.addSet[elem]; !exists {
			s.addSet[elem] = make(map[string]struct{})
		}
		for tag := range otherTags {
			s.addSet[elem][tag] = struct{}{}
		}
	}

	// Union removeSets
	for elem, otherTags := range other.removeSet {
		if _, exists := s.removeSet[elem]; !exists {
			s.removeSet[elem] = make(map[string]struct{})
		}
		for tag := range otherTags {
			s.removeSet[elem][tag] = struct{}{}
		}
	}

	return nil
}

func (s *ORSet[T]) generateTag() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s:%d:%s", s.nodeID, time.Now().UnixNano(), hex.EncodeToString(b[:]))
}
