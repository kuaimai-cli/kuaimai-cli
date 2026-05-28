package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWrite_mapStringString_tableHasOK(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Out: &buf}
	err := p.Success(map[string]string{"message": "配置已初始化", "path": "/tmp/x"})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ok") || !strings.Contains(out, "true") {
		t.Fatalf("table output missing ok:true:\n%s", out)
	}
	if !strings.Contains(out, "message") {
		t.Fatalf("table output missing data fields:\n%s", out)
	}
}

func TestWrite_mapStringString_jsonEnvelope(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Out: &buf}
	if err := p.Success(map[string]string{"message": "done"}); err != nil {
		t.Fatal(err)
	}
	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK {
		t.Fatal("expected ok:true")
	}
	m, ok := env.Data.(map[string]any)
	if !ok || m["message"] != "done" {
		t.Fatalf("unexpected data: %#v", env.Data)
	}
}
