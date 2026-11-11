package paginate

import (
	"fmt"
	"strings"
)

// NullsOrder specifies NULL sorting position in PostgreSQL.
type NullsOrder string

const (
	NullsDefault NullsOrder = ""
	NullsFirst   NullsOrder = "NULLS FIRST"
	NullsLast    NullsOrder = "NULLS LAST"
)

// SQLClause generates a parameterized PostgreSQL ORDER BY, LIMIT, and OFFSET clause.
// startArgIndex specifies the positional index for $1, $2... so you can append to existing WHERE parameters.
//
// Example:
//
//	clause, args := paginate.SQLClause(params, 1)
//	// Returns: "ORDER BY \"created_at\" DESC LIMIT $1 OFFSET $2", []any{10, 0}
func SQLClause(p Params, startArgIndex int) (string, []any) {
	return SQLClauseWithOptions(p, startArgIndex, NullsDefault, "")
}

// SQLClauseWithOptions provides advanced control over NULL positioning and deterministic tie-breaker sorting.
func SQLClauseWithOptions(p Params, startArgIndex int, nulls NullsOrder, tieBreakerCol string) (string, []any) {
	if startArgIndex <= 0 {
		startArgIndex = 1
	}

	quotedCol := quoteIdentifier(p.Sort)
	order := OrderDesc
	if strings.EqualFold(p.Order, OrderAsc) {
		order = OrderAsc
	}

	orderParts := []string{fmt.Sprintf("%s %s", quotedCol, order)}
	if nulls != NullsDefault {
		orderParts[0] = fmt.Sprintf("%s %s %s", quotedCol, order, nulls)
	}

	if tieBreakerCol != "" {
		quotedTie := quoteIdentifier(tieBreakerCol)
		if quotedTie != quotedCol {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", quotedTie, order))
		}
	}

	limitPlaceholder := fmt.Sprintf("$%d", startArgIndex)
	offsetPlaceholder := fmt.Sprintf("$%d", startArgIndex+1)

	clause := fmt.Sprintf("ORDER BY %s LIMIT %s OFFSET %s",
		strings.Join(orderParts, ", "),
		limitPlaceholder,
		offsetPlaceholder,
	)

	args := []any{p.Limit(), p.Offset()}
	return clause, args
}

// BuildCountSQL wraps any valid SQL query in a count subquery.
// GROUP BY, DISTINCT, aur complex JOIN queries me standard SELECT COUNT(*) directly replace karne se fail hota hai.
// Subquery wrapper har case me 100% accurate count return karta hai.
//
// Example:
//
//	countSQL := paginate.BuildCountSQL("SELECT p.id, c.name FROM products p JOIN categories c ON ... WHERE p.price > $1")
//	// Returns: "SELECT COUNT(*) FROM (SELECT p.id, c.name FROM products p JOIN categories c ON ... WHERE p.price > $1) AS _count_subquery"
func BuildCountSQL(baseQuery string) string {
	trimmed := strings.TrimSpace(baseQuery)
	// Remove trailing semicolon if present
	trimmed = strings.TrimSuffix(trimmed, ";")

	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS _count_subquery", trimmed)
}

// CursorWhereClause generates parameterized PostgreSQL condition and ORDER BY for keyset cursor pagination.
// If order is DESC, condition is: "column < $N"
// If order is ASC, condition is: "column > $N"
func CursorWhereClause(column, cursorVal, order string, startArgIndex int) (string, string, []any) {
	if startArgIndex <= 0 {
		startArgIndex = 1
	}

	quotedCol := quoteIdentifier(column)
	cleanOrder := OrderDesc
	operator := "<"

	if strings.EqualFold(order, OrderAsc) {
		cleanOrder = OrderAsc
		operator = ">"
	}

	orderClause := fmt.Sprintf("ORDER BY %s %s", quotedCol, cleanOrder)

	if cursorVal == "" {
		return "", orderClause, nil
	}

	whereClause := fmt.Sprintf("%s %s $%d", quotedCol, operator, startArgIndex)
	return whereClause, orderClause, []any{cursorVal}
}

// CompositeCursorWhereClause generates parameterized PostgreSQL row-value condition and ORDER BY
// for composite keyset cursor pagination (e.g. ORDER BY "created_at" DESC, "id" DESC).
// Generates standard tuple comparison: ("created_at", "id") < ($1, $2)
func CompositeCursorWhereClause(col1, col2 string, val1, val2 any, order string, startArgIndex int) (string, string, []any) {
	if startArgIndex <= 0 {
		startArgIndex = 1
	}

	quotedCol1 := quoteIdentifier(col1)
	quotedCol2 := quoteIdentifier(col2)
	cleanOrder := OrderDesc
	operator := "<"

	if strings.EqualFold(order, OrderAsc) {
		cleanOrder = OrderAsc
		operator = ">"
	}

	orderClause := fmt.Sprintf("ORDER BY %s %s, %s %s", quotedCol1, cleanOrder, quotedCol2, cleanOrder)

	if val1 == nil || val2 == nil {
		return "", orderClause, nil
	}

	whereClause := fmt.Sprintf("(%s, %s) %s ($%d, $%d)", quotedCol1, quotedCol2, operator, startArgIndex, startArgIndex+1)
	return whereClause, orderClause, []any{val1, val2}
}

// quoteIdentifier safely quotes SQL identifiers (e.g. "users"."created_at" or "id")
// to prevent keyword collisions (like user, order, group) in PostgreSQL.
func quoteIdentifier(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `"id"`
	}

	parts := strings.Split(s, ".")
	var quoted []string
	for _, p := range parts {
		cleaned := cleanIdentifier(p)
		if cleaned != "" {
			quoted = append(quoted, fmt.Sprintf(`"%s"`, cleaned))
		}
	}

	if len(quoted) == 0 {
		return `"id"`
	}
	return strings.Join(quoted, ".")
}
