# SPEC: go-queryfilter

## 1. Metadata
- **Module**: `go-queryfilter`
- **Package**: `queryfilter`
- **Location**: `C:\Library\go-queryfilter`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `net/url`, `strings`, `strconv`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **Whitelist-Enforced SQL Query Translation**:
   - Converts standard REST URL query strings into parameterized SQL `WHERE`, `ORDER BY`, `LIMIT`, and `OFFSET` clauses.
   - Enforces strict developer-configured column whitelists (`AllowedFilters`, `AllowedSorts`), blocking SQL injection and unauthorized column access.
2. **Rich Operator Support**:
   - Maps brackets (`field[gte]=10`, `field[in]=a,b,c`, `field[like]=%pattern%`) to standard SQL operators (`=`, `!=`, `>`, `>=`, `<`, `<=`, `LIKE`, `ILIKE`, `IN`).
3. **Sort Direction & Offset Pagination**:
   - Translates prefixed fields (`sort=-created_at,name`) into `ORDER BY created_at DESC, name ASC`.
   - Calculates `LIMIT` and `OFFSET` from `page` and `limit` query parameters with safety caps.

---

## 3. Public API
```go
type Operator string

type Config struct {
    AllowedFilters map[string]bool
    AllowedSorts   map[string]bool
    DefaultSort    string
    DefaultLimit   int
    MaxLimit       int
}

type SQLResult struct {
    WhereClause   string
    OrderByClause string
    Limit         int
    Offset        int
    Args          []any
}

type Parser struct
func New(cfg Config) *Parser
func (p *Parser) Parse(values url.Values, startParamIdx int) (*SQLResult, error)
```
