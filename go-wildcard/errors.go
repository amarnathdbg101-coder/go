package gowildcard

import "errors"

var (
	// ErrPatternSyntax indicates malformed bracket expression in pattern.
	ErrPatternSyntax = errors.New("gowildcard: malformed wildcard pattern")
)
