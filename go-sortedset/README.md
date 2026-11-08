# go-sortedset 🏆

> **High-Performance In-Memory Redis ZSET (Sorted Set) for Go (Go 1.23+ Generics).**  
> Gaming leaderboards, priority scheduling, dynamic rate-limiting, aur time-series ranking ke liye Redis ke ZSET data structure ka pure Go implementation. SkipList (with spans) + Hash Map hybrid ka use karke $O(\log N)$ rank calculation aur range queries provide karta hai.

---

## 🚀 Features

- **SkipList + Hash Table**: $O(1)$ score lookup aur $O(\log N)$ insert, update, delete aur rank calculation.
- **Leaderboard Ready**:
  - `Rank`: 0-based rank from lowest to highest score.
  - `RevRank`: 0-based rank from highest to lowest score (Leaderboard #1, #2, etc.).
- **Range Queries**:
  - `RangeByRank(start, stop)`: Specific ranks ke players fetch karein.
  - `RangeByScore(min, max)`: Score boundaries ke beech ke elements fetch karein.
- **PopMin & PopMax**: Minimum ya Maximum elements ko atomically extract karein.
- **Go 1.23+ Generics**: Kisi bhi `comparable` type ke liye type-safe `Set[M comparable]`.

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-sortedset"
)

func main() {
	// Gaming Leaderboard banayein
	lb := sortedset.New[string]()

	lb.Add("player_arjun", 1540)
	lb.Add("player_vikram", 2800)
	lb.Add("player_rahul", 950)

	// Leaderboard rank check karein (#1 player)
	topRank, _ := lb.RevRank("player_vikram")
	fmt.Printf("Vikram Rank: #%d (0-indexed)\n", topRank) // Output: #0

	// Top 2 players fetch karein
	topPlayers := lb.RangeByRank(0, 1)
	for i, item := range topPlayers {
		fmt.Printf("%d. %s => %.0f pts\n", i+1, item.Member, item.Score)
	}
}
```
