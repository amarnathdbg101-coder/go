package gotree

import (
	"fmt"
	"sync"
)

// Node represents an element in the hierarchical tree.
type Node[T any] struct {
	ID       string     `json:"id"`
	ParentID string     `json:"parent_id,omitempty"`
	Value    T          `json:"value"`
	Children []*Node[T] `json:"children,omitempty"`
}

// Hierarchy manages a thread-safe multi-root category/entity tree.
type Hierarchy[T any] struct {
	nodes map[string]*Node[T]
	roots []*Node[T]
	mu    sync.RWMutex
}

// NewHierarchy initializes an empty hierarchy container.
func NewHierarchy[T any]() *Hierarchy[T] {
	return &Hierarchy[T]{
		nodes: make(map[string]*Node[T]),
		roots: make([]*Node[T], 0),
	}
}

// Add inserts a node into the hierarchy. If parentID is empty, it becomes a root node.
func (h *Hierarchy[T]) Add(id, parentID string, val T) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.nodes[id]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateNode, id)
	}

	node := &Node[T]{
		ID:       id,
		ParentID: parentID,
		Value:    val,
		Children: make([]*Node[T], 0),
	}

	if parentID == "" {
		h.roots = append(h.roots, node)
	} else {
		parent, exists := h.nodes[parentID]
		if !exists {
			return fmt.Errorf("%w: %s", ErrParentNotFound, parentID)
		}
		// Cycle detection
		curr := parent
		for curr != nil {
			if curr.ID == id {
				return ErrCycleDetected
			}
			if curr.ParentID != "" {
				curr = h.nodes[curr.ParentID]
			} else {
				break
			}
		}
		parent.Children = append(parent.Children, node)
	}

	h.nodes[id] = node
	return nil
}

// Get finds a node by ID.
func (h *Hierarchy[T]) Get(id string) (*Node[T], bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	node, exists := h.nodes[id]
	return node, exists
}

// Ancestors returns ordered path from root down to parent of given node.
func (h *Hierarchy[T]) Ancestors(id string) ([]*Node[T], error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	node, exists := h.nodes[id]
	if !exists {
		return nil, ErrNodeNotFound
	}

	var ancestors []*Node[T]
	currParentID := node.ParentID

	for currParentID != "" {
		parent, ok := h.nodes[currParentID]
		if !ok {
			break
		}
		ancestors = append([]*Node[T]{parent}, ancestors...)
		currParentID = parent.ParentID
	}

	return ancestors, nil
}

// Descendants returns a flattened slice of all children under node id.
func (h *Hierarchy[T]) Descendants(id string) ([]*Node[T], error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	node, exists := h.nodes[id]
	if !exists {
		return nil, ErrNodeNotFound
	}

	var descendants []*Node[T]
	var collect func(n *Node[T])
	collect = func(n *Node[T]) {
		for _, child := range n.Children {
			descendants = append(descendants, child)
			collect(child)
		}
	}

	collect(node)
	return descendants, nil
}

// Roots returns top-level root nodes.
func (h *Hierarchy[T]) Roots() []*Node[T] {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.roots
}
