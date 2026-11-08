# SPEC: go-sortedset

## 1. Metadata
- **Module**: `go-sortedset`
- **Package**: `sortedset`
- **Location**: `C:\Library\go-sortedset`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `crypto/rand`, `encoding/binary`, `errors`).

---

## 2. Core Capabilities
1. **Redis ZSET Architecture**:
   - Hybrid data structure pairing a 32-level SkipList with node span pointers and a hash table to achieve $O(1)$ member lookups and $O(\log N)$ additions, removals, and score updates.
2. **Logarithmic Rank & Reverse Rank Calculations**:
   - `Rank(member)` returns 0-based rank from lowest to highest score in $O(\log N)$.
   - `RevRank(member)` returns 0-based rank from highest to lowest score (ideal for gaming & leaderboard engines).
3. **Bounded Range Lookups**:
   - `RangeByRank(start, stop)` and `RangeByScore(min, max)` retrieve slices of items efficiently.
4. **Extreme Element Extractions**:
   - `PopMin()` and `PopMax()` atomically extract endpoints in $O(1)$ and $O(\log N)$ time.
5. **Thread-Safe & Generics Support**:
   - Supports any comparable member type `[M comparable]`. `NewConcurrent[M]()` coordinates multithreaded operations with `sync.RWMutex`.

---

## 3. Public API
```go
type Item[M comparable] struct {
    Member M       `json:"member"`
    Score  float64 `json:"score"`
}

type Set[M comparable] struct
func New[M comparable]() *Set[M]
func NewConcurrent[M comparable]() *Set[M]
func (s *Set[M]) Add(member M, score float64) bool
func (s *Set[M]) Score(member M) (float64, bool)
func (s *Set[M]) Rank(member M) (int, bool)
func (s *Set[M]) RevRank(member M) (int, bool)
func (s *Set[M]) Delete(member M) bool
func (s *Set[M]) Len() int
func (s *Set[M]) RangeByRank(start, stop int) []Item[M]
func (s *Set[M]) RangeByScore(minScore, maxScore float64) []Item[M]
func (s *Set[M]) PopMin() (*Item[M], bool)
func (s *Set[M]) PopMax() (*Item[M], bool)
func (s *Set[M]) Clear()
```
