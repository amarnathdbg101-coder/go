# Technical Specification: go-hashring

## 1. Domain & Purpose
`go-hashring` is a high-performance consistent hash ring with virtual nodes (Ketama algorithm) for distributed sharding, caching, and load balancing in Go. It ensures minimal key remapping ($K/N$) when nodes are added or removed, preventing cache storms and database overload.

## 2. Architecture & Design Principles
- **64-bit Uniform Dispersion Ring**: Combines 64-bit FNV-1a hashing with a Stafford/SplitMix64 bit mixer finalizer to guarantee uniform pseudo-random key dispersion across virtual nodes.
- **Configurable Virtual Nodes**: Maps each physical node to $V$ virtual positions (default 100) to balance workloads evenly across cluster members.
- **Replication Support (`GetN`)**: Retrieves $N$ distinct replica nodes in clockwise circular order for multi-replica storage topologies.
- **Thread Safety**: Fully protected by `sync.RWMutex` for concurrent reads and dynamic node cluster mutations.
- **Zero Heavy Dependencies**: Pure standard library (`hash/fnv`, `sort`, `strconv`, `sync`).

## 3. API Surface
- `type HashRing struct`
- `func New(vnodes int, initialNodes ...string) *HashRing`
- `func (hr *HashRing) Add(node string)`
- `func (hr *HashRing) Remove(node string) error`
- `func (hr *HashRing) Get(key string) (string, error)`
- `func (hr *HashRing) GetN(key string, n int) ([]string, error)`
- `func (hr *HashRing) Nodes() []string`
- `func (hr *HashRing) Len() int`

## 4. Error Handling
- `ErrEmptyRing`: Attempted key lookup on a hash ring containing 0 nodes.
- `ErrNodeNotFound`: Target node does not exist when executing removal.
- `ErrInvalidCount`: Parameter $N \le 0$ requested for replica nodes.

## 5. Verification & Testing
- Unit tests verify deterministic node routing, minimal key migration (~25% upon node removal from a 4-node cluster), distinct replica ring iteration, empty ring errors, and concurrent thread-safe mutations.
