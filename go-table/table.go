package table

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Alignment defines the text justification inside a table cell.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Table formats tabular data into aligned terminal tables, markdown tables, or CSV outputs.
type Table struct {
	headers []string
	rows    [][]string
	aligns  map[int]Alignment
}

// New creates a new Table with the specified header titles.
func New(headers ...string) *Table {
	return &Table{
		headers: headers,
		aligns:  make(map[int]Alignment),
	}
}

// AddRow appends a row of cells to the table, converting each item to string.
func (t *Table) AddRow(cells ...any) *Table {
	row := make([]string, len(cells))
	for i, c := range cells {
		row[i] = fmt.Sprintf("%v", c)
	}
	t.rows = append(t.rows, row)
	return t
}

// SetAlignment configures text alignment for a specific column index (0-based).
func (t *Table) SetAlignment(colIndex int, align Alignment) *Table {
	t.aligns[colIndex] = align
	return t
}

// Render produces a formatted ASCII grid table.
func (t *Table) Render() string {
	widths := t.calcColWidths()
	numCols := len(widths)
	if numCols == 0 {
		return ""
	}

	var sb strings.Builder
	border := t.buildBorder(widths)

	sb.WriteString(border)
	sb.WriteByte('\n')

	// Header
	if len(t.headers) > 0 {
		sb.WriteString(t.renderRow(t.headers, widths))
		sb.WriteByte('\n')
		sb.WriteString(border)
		sb.WriteByte('\n')
	}

	// Rows
	for _, r := range t.rows {
		sb.WriteString(t.renderRow(r, widths))
		sb.WriteByte('\n')
	}

	sb.WriteString(border)
	return sb.String()
}

// RenderMarkdown produces a GitHub-flavored Markdown table.
func (t *Table) RenderMarkdown() string {
	widths := t.calcColWidths()
	numCols := len(widths)
	if numCols == 0 {
		return ""
	}

	var sb strings.Builder

	// Header row
	sb.WriteString(t.renderRow(t.headers, widths))
	sb.WriteByte('\n')

	// Divider row with alignment indicators
	sb.WriteString("|")
	for i, w := range widths {
		align := t.aligns[i]
		switch align {
		case AlignCenter:
			sb.WriteString(" :" + strings.Repeat("-", max(1, w)) + ": |")
		case AlignRight:
			sb.WriteString(" " + strings.Repeat("-", max(1, w+1)) + ": |")
		default: // AlignLeft
			sb.WriteString(" :" + strings.Repeat("-", max(1, w+1)) + " |")
		}
	}
	sb.WriteByte('\n')

	// Rows
	for _, r := range t.rows {
		sb.WriteString(t.renderRow(r, widths))
		sb.WriteByte('\n')
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// RenderCSV produces comma-separated values.
func (t *Table) RenderCSV() string {
	var sb strings.Builder

	if len(t.headers) > 0 {
		sb.WriteString(t.formatCSVRow(t.headers))
		sb.WriteByte('\n')
	}

	for _, r := range t.rows {
		sb.WriteString(t.formatCSVRow(r))
		sb.WriteByte('\n')
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

func (t *Table) formatCSVRow(cells []string) string {
	var escaped []string
	for _, c := range cells {
		if strings.ContainsAny(c, ",\"\n\r") {
			escaped = append(escaped, `"`+strings.ReplaceAll(c, `"`, `""`)+`"`)
		} else {
			escaped = append(escaped, c)
		}
	}
	return strings.Join(escaped, ",")
}

func (t *Table) renderRow(cells []string, widths []int) string {
	var sb strings.Builder
	sb.WriteString("|")

	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		align := t.aligns[i]
		sb.WriteString(" ")
		sb.WriteString(pad(cell, w, align))
		sb.WriteString(" |")
	}

	return sb.String()
}

func (t *Table) buildBorder(widths []int) string {
	var sb strings.Builder
	sb.WriteString("+")
	for _, w := range widths {
		sb.WriteString(strings.Repeat("-", w+2))
		sb.WriteString("+")
	}
	return sb.String()
}

func (t *Table) calcColWidths() []int {
	numCols := len(t.headers)
	for _, r := range t.rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}

	widths := make([]int, numCols)
	for i, h := range t.headers {
		w := utf8.RuneCountInString(h)
		if w > widths[i] {
			widths[i] = w
		}
	}

	for _, r := range t.rows {
		for i, c := range r {
			w := utf8.RuneCountInString(c)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	return widths
}

func pad(text string, width int, align Alignment) string {
	textLen := utf8.RuneCountInString(text)
	if textLen >= width {
		return text
	}

	diff := width - textLen
	switch align {
	case AlignRight:
		return strings.Repeat(" ", diff) + text
	case AlignCenter:
		left := diff / 2
		right := diff - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default: // AlignLeft
		return text + strings.Repeat(" ", diff)
	}
}
