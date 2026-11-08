# go-queryfilter 🔎

> **Safe REST URL Query to SQL WHERE & ORDER BY Parameterized Builder for Go (Go 1.23+).**  
> REST APIs me `?status=active&age[gte]=18&sort=-created_at,name&page=2` jaise URL query strings ko safe parameterized PostgreSQL SQL clauses (`status = $1 AND age >= $2 ORDER BY created_at DESC, name ASC`) me translate karta hai. Strict whitelist protection se SQL injection 100% impossible hota hai.

---

## 🚀 Features

- **Whitelist Enforcement**: Sirf vahi columns query/sort ho sakte hain jo `AllowedFilters` aur `AllowedSorts` me explicit defined hain.
- **Rich Operator Mappings**:
  - `[eq]`, `[ne]`: `=`, `!=`
  - `[gt]`, `[gte]`, `[lt]`, `[lte]`: `>`, `>=`, `<`, `<=`
  - `[like]`, `[ilike]`: `LIKE`, `ILIKE`
  - `[in]`: `IN ($1, $2, $3)` (Comma-separated auto-expansion)
- **Automatic Sort Direction**: `-created_at` ko `created_at DESC` aur `name` ko `name ASC` banata hai.
- **Page & Offset Calculation**: `page=2&limit=20` ko automatically `LIMIT 20 OFFSET 20` me compute karta hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-queryfilter"
	"net/url"
)

func main() {
	parser := queryfilter.New(queryfilter.Config{
		AllowedFilters: map[string]bool{
			"status": true,
			"price":  true,
		},
		AllowedSorts: map[string]bool{
			"created_at": true,
		},
		DefaultLimit: 20,
	})

	// User HTTP URL Query: ?status=active&price[lte]=500&sort=-created_at&page=1
	queryValues, _ := url.ParseQuery("status=active&price[lte]=500&sort=-created_at&page=1")

	res, err := parser.Parse(queryValues, 1)
	if err != nil {
		panic(err)
	}

	fmt.Printf("SQL WHERE: %s\n", res.WhereClause)
	fmt.Printf("SQL ORDER: %s\n", res.OrderByClause)
	fmt.Printf("Arguments: %v\n", res.Args)
	fmt.Printf("Limit: %d, Offset: %d\n", res.Limit, res.Offset)
}
```
