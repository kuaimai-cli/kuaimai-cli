package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteCSV_listRows(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatCSV, Out: &buf}
	err := p.Success([]map[string]any{
		{"title": "A", "pageNo": 1},
		{"title": "B", "pageNo": 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header+rows, got:\n%s", buf.String())
	}
	if !strings.Contains(lines[0], "title") {
		t.Fatalf("missing header: %s", lines[0])
	}
}

func TestWriteNDJSON_listRows(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatNDJSON, Out: &buf}
	if err := p.Success([]map[string]any{{"id": 1}, {"id": 2}}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), buf.String())
	}
	var row map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &row); err != nil {
		t.Fatal(err)
	}
}

func TestWriteCSV_errorUsesEnvelope(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatCSV, Out: &buf}
	if err := p.Fail("bad", "hint"); err != nil {
		t.Fatal(err)
	}
	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("expected json envelope on error: %s", buf.String())
	}
	if env.OK {
		t.Fatal("expected ok:false")
	}
}
