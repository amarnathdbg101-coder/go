package goevents

import (
	"strings"
)

// MatchTopic checks whether a topic matches a subscription pattern.
// Patterns support dot-separated namespaces with wildcards:
//   - "*" matches exactly one segment (e.g. "order.*" matches "order.created" but not "order.item.added")
//   - "**" or ">" matches zero or more segments (e.g. "order.**" matches "order.created" and "order.item.added")
//   - Exact string match when no wildcards are present.
func MatchTopic(pattern, topic string) bool {
	if pattern == "" || topic == "" {
		return false
	}
	if pattern == "**" || pattern == ">" {
		return true
	}
	if pattern == topic {
		return true
	}

	pParts := strings.Split(pattern, ".")
	tParts := strings.Split(topic, ".")

	pLen := len(pParts)
	tLen := len(tParts)

	pIdx := 0
	tIdx := 0

	for pIdx < pLen && tIdx < tLen {
		pToken := pParts[pIdx]

		if pToken == "**" || pToken == ">" {
			// If multi-level wildcard is at the end, matches everything remaining
			if pIdx == pLen-1 {
				return true
			}
			// Look ahead: find next non-wildcard segment in topic
			nextPToken := pParts[pIdx+1]
			matched := false
			for tIdx < tLen {
				if tParts[tIdx] == nextPToken || nextPToken == "*" {
					matched = true
					break
				}
				tIdx++
			}
			if !matched {
				return false
			}
			pIdx++
			continue
		}

		if pToken == "*" || pToken == tParts[tIdx] {
			pIdx++
			tIdx++
			continue
		}

		return false
	}

	// Handle trailing multi-level wildcard
	if pIdx == pLen-1 && (pParts[pIdx] == "**" || pParts[pIdx] == ">") {
		return true
	}

	return pIdx == pLen && tIdx == tLen
}
