package goversion

import (
	"fmt"
	"strings"
)

type opType int

const (
	opEqual opType = iota
	opNotEqual
	opGreaterThan
	opGreaterThanOrEqual
	opLessThan
	opLessThanOrEqual
)

type singleRule struct {
	op      opType
	version *Version
}

func (r singleRule) check(v *Version) bool {
	cmp := v.Compare(r.version)
	switch r.op {
	case opEqual:
		return cmp == 0
	case opNotEqual:
		return cmp != 0
	case opGreaterThan:
		return cmp > 0
	case opGreaterThanOrEqual:
		return cmp >= 0
	case opLessThan:
		return cmp < 0
	case opLessThanOrEqual:
		return cmp <= 0
	}
	return false
}

// Constraint evaluates whether a version satisfies a set of range criteria.
type Constraint struct {
	rules []singleRule
	expr  string
}

// ParseConstraint parses version constraint expressions (e.g. ">= 1.0.0, < 2.0.0", "^1.2.0", "~1.0.0").
func ParseConstraint(expr string) (*Constraint, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" || expr == "*" {
		return &Constraint{expr: expr}, nil
	}

	c := &Constraint{expr: expr}
	clauses := strings.Split(expr, ",")

	for _, clause := range clauses {
		clause = strings.TrimSpace(clause)
		if clause == "" {
			continue
		}

		// Handle Caret range: ^1.2.3
		if strings.HasPrefix(clause, "^") {
			verStr := strings.TrimSpace(strings.TrimPrefix(clause, "^"))
			v, err := Parse(verStr)
			if err != nil {
				return nil, err
			}
			c.rules = append(c.rules, singleRule{op: opGreaterThanOrEqual, version: v})

			// Upper bound
			var upper *Version
			if v.Major > 0 {
				upper = &Version{Major: v.Major + 1, Minor: 0, Patch: 0}
			} else if v.Minor > 0 {
				upper = &Version{Major: 0, Minor: v.Minor + 1, Patch: 0}
			} else {
				upper = &Version{Major: 0, Minor: 0, Patch: v.Patch + 1}
			}
			c.rules = append(c.rules, singleRule{op: opLessThan, version: upper})
			continue
		}

		// Handle Tilde range: ~1.2.3
		if strings.HasPrefix(clause, "~") {
			verStr := strings.TrimSpace(strings.TrimPrefix(clause, "~"))
			v, err := Parse(verStr)
			if err != nil {
				return nil, err
			}
			c.rules = append(c.rules, singleRule{op: opGreaterThanOrEqual, version: v})
			upper := &Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
			c.rules = append(c.rules, singleRule{op: opLessThan, version: upper})
			continue
		}

		// Standard comparison operators
		var op opType
		var verPart string

		if strings.HasPrefix(clause, ">=") {
			op = opGreaterThanOrEqual
			verPart = strings.TrimPrefix(clause, ">=")
		} else if strings.HasPrefix(clause, "<=") {
			op = opLessThanOrEqual
			verPart = strings.TrimPrefix(clause, "<=")
		} else if strings.HasPrefix(clause, "!=") {
			op = opNotEqual
			verPart = strings.TrimPrefix(clause, "!=")
		} else if strings.HasPrefix(clause, ">") {
			op = opGreaterThan
			verPart = strings.TrimPrefix(clause, ">")
		} else if strings.HasPrefix(clause, "<") {
			op = opLessThan
			verPart = strings.TrimPrefix(clause, "<")
		} else if strings.HasPrefix(clause, "=") {
			op = opEqual
			verPart = strings.TrimPrefix(clause, "=")
		} else {
			op = opEqual
			verPart = clause
		}

		verPart = strings.TrimSpace(verPart)
		v, err := Parse(verPart)
		if err != nil {
			return nil, fmt.Errorf("%w: in clause %q", ErrInvalidConstraint, clause)
		}

		c.rules = append(c.rules, singleRule{op: op, version: v})
	}

	return c, nil
}

// Check evaluates whether the provided version satisfies all constraint rules.
func (c *Constraint) Check(v *Version) bool {
	if len(c.rules) == 0 {
		return true // matches wildcard '*' or empty
	}

	for _, r := range c.rules {
		if !r.check(v) {
			return false
		}
	}
	return true
}

// CheckString parses the version string and evaluates it against the constraint.
func (c *Constraint) CheckString(vStr string) bool {
	v, err := Parse(vStr)
	if err != nil {
		return false
	}
	return c.Check(v)
}
