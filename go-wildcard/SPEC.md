# Technical Specification: go-wildcard

## 1. Domain & Purpose
`go-wildcard` is an ultra-fast, zero-allocation glob and wildcard pattern matching library for Go. While regular expressions (`regexp`) introduce heavy CPU compilation overhead and heap allocations for simple glob searches, `go-wildcard` provides a non-recursive, linear-time $O(n)$ two-pointer algorithm.

## 2. Architecture & Design Principles
- **Linear-Time Two-Pointer Scanning**: Iteratively tracks characters and asterisk backtracking indices (`starIdx`, `matchIdx`), completely avoiding recursion or regex state machine compilations.
- **Tokens**:
  - `*`: Matches zero or more arbitrary characters.
  - `?`: Matches exactly one character.
- **Case-Insensitive Folding**: `MatchFold` handles case-folded matching across Unicode characters.
- **Zero Allocations**: Evaluates patterns inline with 0 B/op heap allocation.
- **Zero Heavy Dependencies**: Pure Go standard library (`strings`, `unicode`).

## 3. API Surface
- `func Match(pattern, text string) bool`
- `func MatchFold(pattern, text string) bool`
- `func MatchAny(patterns []string, text string) bool`
- `func Filter(pattern string, items []string) []string`

## 4. Verification & Testing
- Unit tests verify multi-asterisk backtracking (`*order*paid*`), single character `?` positioning, case-insensitive matching (`*.PNG` -> `photo.png`), and slice filtering.
