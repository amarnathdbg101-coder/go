# Technical Specification: go-tree

## 1. Domain & Purpose
`go-tree` provides generic, thread-safe Prefix Trie and Hierarchical Category Tree data structures in pure Go. It delivers rapid prefix matching and autocompletion along with hierarchical tree management for catalog categories, organizational charts, breadcrumbs, and taxonomies.

## 2. Architecture & Design Principles
- **Generic Payload `T`**: Both Trie and Hierarchy nodes support arbitrary user payloads without type assertions.
- **Thread Safety**: Concurrent read/write protection with `sync.RWMutex`.
- **Hierarchical Path Computations**: Breadcrumb trail resolution (`Ancestors`) and full recursive sub-tree flattening (`Descendants`).
- **Integrity Validation**: Guarded against duplicate IDs, dangling parents, and cyclic loops.
- **Zero Heavy Dependencies**: Pure Go standard library.

## 3. API Surface
- `type Trie[T any] struct`
- `func NewTrie[T any]() *Trie[T]`
- `func (t *Trie[T]) Insert(key string, val T)`
- `func (t *Trie[T]) Search(key string) (T, bool)`
- `func (t *Trie[T]) StartsWith(prefix string) []T`
- `func (t *Trie[T]) Delete(key string) bool`
- `type Node[T any] struct { ID, ParentID string; Value T; Children []*Node[T] }`
- `type Hierarchy[T any] struct`
- `func NewHierarchy[T any]() *Hierarchy[T]`
- `func (h *Hierarchy[T]) Add(id, parentID string, val T) error`
- `func (h *Hierarchy[T]) Get(id string) (*Node[T], bool)`
- `func (h *Hierarchy[T]) Ancestors(id string) ([]*Node[T], error)`
- `func (h *Hierarchy[T]) Descendants(id string) ([]*Node[T], error)`
- `func (h *Hierarchy[T]) Roots() []*Node[T]`

## 4. Error Handling
- `ErrNodeNotFound`: Query on non-existent node ID.
- `ErrDuplicateNode`: Insertion with duplicate ID.
- `ErrParentNotFound`: Parent ID does not exist in hierarchy.
- `ErrCycleDetected`: Cyclic link detected.

## 5. Verification & Testing
- Unit tests verify Trie insertion, prefix matching (`StartsWith`), clean deletion, category hierarchy construction, breadcrumb path computation, and descendant flattening.
