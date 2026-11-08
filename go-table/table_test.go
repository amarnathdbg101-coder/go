package table

import (
	"strings"
	"testing"
)

func TestTable_RenderASCII(t *testing.T) {
	tbl := New("ID", "Product", "Price")
	tbl.AddRow(1, "Keyboard", 1200)
	tbl.AddRow(2, "Mouse", 650.50)
	tbl.AddRow(3, "Monitor", 9999)

	tbl.SetAlignment(0, AlignRight)
	tbl.SetAlignment(2, AlignRight)

	out := tbl.Render()
	if !strings.Contains(out, "Keyboard") || !strings.Contains(out, "9999") {
		t.Errorf("rendered table missing rows: \n%s", out)
	}

	// Verify border formatting
	lines := strings.Split(out, "\n")
	if len(lines) < 7 {
		t.Errorf("expected at least 7 lines in grid table, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "+") {
		t.Errorf("expected border prefix '+', got %s", lines[0])
	}
}

func TestTable_RenderMarkdown(t *testing.T) {
	tbl := New("Name", "Role", "Score")
	tbl.AddRow("Amit", "Admin", 98)
	tbl.AddRow("Priya", "Editor", 85)

	tbl.SetAlignment(0, AlignLeft)
	tbl.SetAlignment(1, AlignCenter)
	tbl.SetAlignment(2, AlignRight)

	md := tbl.RenderMarkdown()

	if !strings.Contains(md, "| Name") || !strings.Contains(md, "| Amit") {
		t.Errorf("markdown table mismatch: \n%s", md)
	}

	// Verify alignment indicators
	if !strings.Contains(md, ":---") || !strings.Contains(md, ":-") || !strings.Contains(md, "---:") {
		t.Errorf("expected markdown alignment colons, got: \n%s", md)
	}
}

func TestTable_RenderCSV(t *testing.T) {
	tbl := New("City", "Description")
	tbl.AddRow("Mumbai", "Financial capital, coastal")
	tbl.AddRow("Delhi", "National capital")

	csv := tbl.RenderCSV()
	expectedEscaped := `"Financial capital, coastal"`
	if !strings.Contains(csv, expectedEscaped) {
		t.Errorf("expected quoted CSV field for comma, got: \n%s", csv)
	}
}

func TestTable_Unicode(t *testing.T) {
	tbl := New("ID", "Name")
	tbl.AddRow(1, "नमस्ते")
	tbl.AddRow(2, "World")

	out := tbl.Render()
	if !strings.Contains(out, "नमस्ते") {
		t.Errorf("expected unicode name in table")
	}
}
