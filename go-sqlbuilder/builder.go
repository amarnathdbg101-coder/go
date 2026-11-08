package gosqlbuilder

import (
	"fmt"
	"strconv"
	"strings"
)

// SelectBuilder constructs parameterized SELECT queries with dynamic WHERE filters.
type SelectBuilder struct {
	table     string
	columns   []string
	joins     []string
	wheres    []string
	args      []any
	orderBys  []string
	limit     int
	offset    int
	argOffset int
}

// Select initializes a SelectBuilder with requested columns.
func Select(columns ...string) *SelectBuilder {
	return &SelectBuilder{
		columns: columns,
		limit:   -1,
		offset:  -1,
	}
}

// From sets the target table.
func (b *SelectBuilder) From(table string) *SelectBuilder {
	b.table = strings.TrimSpace(table)
	return b
}

// Join appends a JOIN clause (e.g. "LEFT JOIN", "users", "users.id = orders.user_id").
func (b *SelectBuilder) Join(joinType, table, on string) *SelectBuilder {
	b.joins = append(b.joins, fmt.Sprintf("%s %s ON %s", strings.TrimSpace(joinType), strings.TrimSpace(table), strings.TrimSpace(on)))
	return b
}

// Where appends a condition replacing question marks '?' with PostgreSQL parameterized placeholders ($1, $2, ...).
func (b *SelectBuilder) Where(condition string, args ...any) *SelectBuilder {
	if strings.TrimSpace(condition) == "" {
		return b
	}

	formattedCond := b.replacePlaceholders(condition, len(args))
	b.wheres = append(b.wheres, formattedCond)
	b.args = append(b.args, args...)
	return b
}

// WhereIn appends an "IN ($1, $2, ...)" condition.
func (b *SelectBuilder) WhereIn(column string, args ...any) *SelectBuilder {
	if len(args) == 0 {
		return b
	}

	var placeholders []string
	for range args {
		b.argOffset++
		placeholders = append(placeholders, "$"+strconv.Itoa(b.argOffset))
	}

	cond := fmt.Sprintf("%s IN (%s)", strings.TrimSpace(column), strings.Join(placeholders, ", "))
	b.wheres = append(b.wheres, cond)
	b.args = append(b.args, args...)
	return b
}

// OrderBy appends an ORDER BY clause. Strictly validates direction to prevent SQL injection.
func (b *SelectBuilder) OrderBy(column, direction string) *SelectBuilder {
	dir := strings.ToUpper(strings.TrimSpace(direction))
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}
	b.orderBys = append(b.orderBys, fmt.Sprintf("%s %s", strings.TrimSpace(column), dir))
	return b
}

// Limit sets the query LIMIT.
func (b *SelectBuilder) Limit(limit int) *SelectBuilder {
	b.limit = limit
	return b
}

// Offset sets the query OFFSET.
func (b *SelectBuilder) Offset(offset int) *SelectBuilder {
	b.offset = offset
	return b
}

func (b *SelectBuilder) replacePlaceholders(cond string, count int) string {
	var sb strings.Builder
	argIdx := 0

	for i := 0; i < len(cond); i++ {
		if cond[i] == '?' && argIdx < count {
			b.argOffset++
			sb.WriteString("$")
			sb.WriteString(strconv.Itoa(b.argOffset))
			argIdx++
		} else {
			sb.WriteByte(cond[i])
		}
	}
	return sb.String()
}

// Build generates the final PostgreSQL SQL statement and parameterized arguments.
func (b *SelectBuilder) Build() (string, []any, error) {
	if b.table == "" {
		return "", nil, ErrNoTable
	}

	cols := "*"
	if len(b.columns) > 0 {
		cols = strings.Join(b.columns, ", ")
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(cols)
	sb.WriteString(" FROM ")
	sb.WriteString(b.table)

	for _, j := range b.joins {
		sb.WriteString(" ")
		sb.WriteString(j)
	}

	if len(b.wheres) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.wheres, " AND "))
	}

	if len(b.orderBys) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(strings.Join(b.orderBys, ", "))
	}

	if b.limit >= 0 {
		b.argOffset++
		sb.WriteString(" LIMIT $")
		sb.WriteString(strconv.Itoa(b.argOffset))
		b.args = append(b.args, b.limit)
	}

	if b.offset >= 0 {
		b.argOffset++
		sb.WriteString(" OFFSET $")
		sb.WriteString(strconv.Itoa(b.argOffset))
		b.args = append(b.args, b.offset)
	}

	return sb.String(), b.args, nil
}

// InsertBuilder builds parameterized INSERT statements.
type InsertBuilder struct {
	table      string
	columns    []string
	values     []any
	returning  []string
	rowLengths []int
}

// InsertInto initializes an InsertBuilder for the target table.
func InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{table: strings.TrimSpace(table)}
}

// Columns specifies the destination column names.
func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	b.columns = columns
	return b
}

// Values appends a row of values.
func (b *InsertBuilder) Values(values ...any) *InsertBuilder {
	b.values = append(b.values, values...)
	b.rowLengths = append(b.rowLengths, len(values))
	return b
}

// Returning appends RETURNING clause columns.
func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.returning = columns
	return b
}

// Build generates the parameterized INSERT statement.
func (b *InsertBuilder) Build() (string, []any, error) {
	if b.table == "" {
		return "", nil, ErrNoTable
	}
	if len(b.columns) == 0 {
		return "", nil, ErrNoColumns
	}
	if len(b.values) == 0 {
		return "", nil, ErrNoValues
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(b.columns, ", "))
	sb.WriteString(") VALUES ")

	argIdx := 1
	var rowPlaceholders []string
	valIdx := 0

	for _, length := range b.rowLengths {
		if length != len(b.columns) {
			return "", nil, ErrMismatchedValues
		}
		var phs []string
		for i := 0; i < length; i++ {
			phs = append(phs, "$"+strconv.Itoa(argIdx))
			argIdx++
			valIdx++
		}
		rowPlaceholders = append(rowPlaceholders, "("+strings.Join(phs, ", ")+")")
	}

	sb.WriteString(strings.Join(rowPlaceholders, ", "))

	if len(b.returning) > 0 {
		sb.WriteString(" RETURNING ")
		sb.WriteString(strings.Join(b.returning, ", "))
	}

	return sb.String(), b.values, nil
}

// UpdateBuilder builds parameterized UPDATE statements.
type UpdateBuilder struct {
	table     string
	sets      []string
	wheres    []string
	args      []any
	argOffset int
	returning []string
}

// Update initializes an UpdateBuilder.
func Update(table string) *UpdateBuilder {
	return &UpdateBuilder{table: strings.TrimSpace(table)}
}

// Set appends a column = value assignment.
func (b *UpdateBuilder) Set(column string, value any) *UpdateBuilder {
	b.argOffset++
	b.sets = append(b.sets, fmt.Sprintf("%s = $%d", strings.TrimSpace(column), b.argOffset))
	b.args = append(b.args, value)
	return b
}

// Where appends an update condition.
func (b *UpdateBuilder) Where(condition string, args ...any) *UpdateBuilder {
	for i := 0; i < len(condition); i++ {
		if condition[i] == '?' {
			b.argOffset++
			condition = condition[:i] + "$" + strconv.Itoa(b.argOffset) + condition[i+1:]
		}
	}
	b.wheres = append(b.wheres, condition)
	b.args = append(b.args, args...)
	return b
}

// Returning appends RETURNING clause.
func (b *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
	b.returning = columns
	return b
}

// Build generates the parameterized UPDATE statement.
func (b *UpdateBuilder) Build() (string, []any, error) {
	if b.table == "" {
		return "", nil, ErrNoTable
	}
	if len(b.sets) == 0 {
		return "", nil, ErrNoSetClauses
	}

	var sb strings.Builder
	sb.WriteString("UPDATE ")
	sb.WriteString(b.table)
	sb.WriteString(" SET ")
	sb.WriteString(strings.Join(b.sets, ", "))

	if len(b.wheres) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.wheres, " AND "))
	}

	if len(b.returning) > 0 {
		sb.WriteString(" RETURNING ")
		sb.WriteString(strings.Join(b.returning, ", "))
	}

	return sb.String(), b.args, nil
}

// DeleteBuilder builds parameterized DELETE statements.
type DeleteBuilder struct {
	table     string
	wheres    []string
	args      []any
	argOffset int
}

// DeleteFrom initializes a DeleteBuilder.
func DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{table: strings.TrimSpace(table)}
}

// Where appends a delete condition.
func (b *DeleteBuilder) Where(condition string, args ...any) *DeleteBuilder {
	for i := 0; i < len(condition); i++ {
		if condition[i] == '?' {
			b.argOffset++
			condition = condition[:i] + "$" + strconv.Itoa(b.argOffset) + condition[i+1:]
		}
	}
	b.wheres = append(b.wheres, condition)
	b.args = append(b.args, args...)
	return b
}

// Build generates the DELETE statement.
func (b *DeleteBuilder) Build() (string, []any, error) {
	if b.table == "" {
		return "", nil, ErrNoTable
	}

	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	sb.WriteString(b.table)

	if len(b.wheres) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.wheres, " AND "))
	}

	return sb.String(), b.args, nil
}
