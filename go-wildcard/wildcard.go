package gowildcard

import (
	"strings"
	"unicode"
)

// Match returns true if text satisfies the wildcard pattern.
// Supported tokens:
//   '*' matches any sequence of zero or more characters.
//   '?' matches exactly one single character.
// Zero heap allocations.
func Match(pattern, text string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == text {
		return true
	}

	pRunes := []rune(pattern)
	tRunes := []rune(text)

	pIdx, tIdx := 0, 0
	starIdx, matchIdx := -1, 0

	pLen := len(pRunes)
	tLen := len(tRunes)

	for tIdx < tLen {
		if pIdx < pLen && (pRunes[pIdx] == '?' || pRunes[pIdx] == tRunes[tIdx]) {
			pIdx++
			tIdx++
		} else if pIdx < pLen && pRunes[pIdx] == '*' {
			starIdx = pIdx
			matchIdx = tIdx
			pIdx++
		} else if starIdx != -1 {
			pIdx = starIdx + 1
			matchIdx++
			tIdx = matchIdx
		} else {
			return false
		}
	}

	for pIdx < pLen && pRunes[pIdx] == '*' {
		pIdx++
	}

	return pIdx == pLen
}

// MatchFold returns true if text satisfies pattern using case-insensitive matching.
func MatchFold(pattern, text string) bool {
	if pattern == "*" {
		return true
	}
	if strings.EqualFold(pattern, text) {
		return true
	}

	pRunes := []rune(pattern)
	tRunes := []rune(text)

	pIdx, tIdx := 0, 0
	starIdx, matchIdx := -1, 0

	pLen := len(pRunes)
	tLen := len(tRunes)

	for tIdx < tLen {
		matchChar := pIdx < pLen && (pRunes[pIdx] == '?' || unicode.ToLower(pRunes[pIdx]) == unicode.ToLower(tRunes[tIdx]))
		if matchChar {
			pIdx++
			tIdx++
		} else if pIdx < pLen && pRunes[pIdx] == '*' {
			starIdx = pIdx
			matchIdx = tIdx
			pIdx++
		} else if starIdx != -1 {
			pIdx = starIdx + 1
			matchIdx++
			tIdx = matchIdx
		} else {
			return false
		}
	}

	for pIdx < pLen && pRunes[pIdx] == '*' {
		pIdx++
	}

	return pIdx == pLen
}

// MatchAny returns true if text matches any of the given patterns.
func MatchAny(patterns []string, text string) bool {
	for _, p := range patterns {
		if Match(p, text) {
			return true
		}
	}
	return false
}

// Filter returns all items from slice that match the given wildcard pattern.
func Filter(pattern string, items []string) []string {
	var result []string
	for _, it := range items {
		if Match(pattern, it) {
			result = append(result, it)
		}
	}
	return result
}
