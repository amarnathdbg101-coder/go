# Technical Specification: go-table

## 1. Domain & Purpose
`go-table` is a lightweight, zero-dependency ASCII, Markdown, and CSV tabular data formatting engine for Go. It structures dynamic data into aligned terminal tables, GitHub-flavored Markdown tables, and CSV exports with column text alignment and Unicode-aware character width calculations.

## 2. Architecture & Design Principles
- **Multi-Format Renderers**:
  - `Render()`: Clean ASCII boxed border grids (`+---+-+`).
  - `RenderMarkdown()`: GitHub-flavored Markdown with alignment markers (`:---`, `:---:`, `---:`).
  - `RenderCSV()`: RFC 4180 comma-separated values with automatic quote escaping.
- **Unicode Width Alignment**: Calculates column widths using `utf8.RuneCountInString` to prevent table border misalignment with non-ASCII or multi-byte UTF-8 scripts.
- **Configurable Justification**: Left, Center, and Right cell alignment per column.
- **Zero Heavy Dependencies**: Pure standard library (`strings`, `fmt`, `unicode/utf8`).

## 3. API Surface
- `type Alignment int`
- `type Table struct`
- `func New(headers ...string) *Table`
- `func (t *Table) AddRow(cells ...any) *Table`
- `func (t *Table) SetAlignment(colIndex int, align Alignment) *Table`
- `func (t *Table) Render() string`
- `func (t *Table) RenderMarkdown() string`
- `func (t *Table) RenderCSV() string`

## 4. Error Handling
- Safe default handling on empty rows with graceful empty strings.

## 5. Verification & Testing
- Unit tests verify ASCII box rendering, Markdown alignment dividers, CSV field quoting, heterogeneous cell types, and Unicode string alignment.
