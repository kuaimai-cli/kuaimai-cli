package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Format is the user-selected output style.
type Format string

const (
	FormatTable  Format = "table"
	FormatJSON   Format = "json"
	FormatCSV    Format = "csv"
	FormatNDJSON Format = "ndjson"
)

// ParseFormat validates a user-facing output format name.
func ParseFormat(raw string) (Format, bool) {
	switch Format(strings.TrimSpace(strings.ToLower(raw))) {
	case FormatTable, FormatJSON, FormatCSV, FormatNDJSON:
		return Format(strings.TrimSpace(strings.ToLower(raw))), true
	default:
		return "", false
	}
}

// Envelope is the fixed CLI response structure (stdout).
type Envelope struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Hint  string `json:"hint,omitempty"`
}

// Printer renders envelopes to stdout.
type Printer struct {
	Format  Format
	Out     io.Writer
	Color   bool
}

// NewPrinter creates a printer defaulting to table on stdout.
func NewPrinter(format Format) *Printer {
	if format == "" {
		format = FormatTable
	}
	return &Printer{Format: format, Out: os.Stdout, Color: true}
}

// SetColor enables or disables colored terminal output.
func (p *Printer) SetColor(enabled bool) {
	p.Color = enabled
}

// Success writes a successful envelope.
func (p *Printer) Success(data any) error {
	return p.Write(Envelope{OK: true, Data: data})
}

// Fail writes a failed envelope (still exits 0 from output layer; caller sets exit code).
func (p *Printer) Fail(errMsg, hint string) error {
	return p.Write(Envelope{OK: false, Error: errMsg, Hint: hint})
}

// Write serializes the envelope to stdout.
func (p *Printer) Write(env Envelope) error {
	switch p.Format {
	case FormatJSON:
		enc := json.NewEncoder(p.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(env)
	case FormatCSV:
		return writeCSV(p.Out, env)
	case FormatNDJSON:
		return writeNDJSON(p.Out, env)
	default:
		return p.writeTable(env)
	}
}

func (p *Printer) tableStyle() table.Style {
	if !p.Color {
		return table.StyleDefault
	}
	return table.StyleColoredBright
}

func (p *Printer) okLine(ok bool) {
	line := "ok: false"
	color := text.FgRed
	if ok {
		line = "ok: true"
		color = text.FgGreen
	}
	if p.Color {
		fmt.Fprintln(p.Out, color.Sprint(line))
		return
	}
	fmt.Fprintln(p.Out, line)
}

func (p *Printer) writeTable(env Envelope) error {
	if !env.OK {
		return p.renderKVTable(map[string]string{
			"ok":    "false",
			"error": env.Error,
			"hint":  env.Hint,
		}, true)
	}

	if rows, ok := tabularRows(env.Data); ok {
		p.okLine(true)
		tw := table.NewWriter()
		tw.SetOutputMirror(p.Out)
		tw.SetStyle(p.tableStyle())
		if len(rows) > 0 {
			tw.AppendHeader(rows[0])
			for i := 1; i < len(rows); i++ {
				tw.AppendRow(rows[i])
			}
		}
		tw.Render()
		return nil
	}

	if kv := flattenData(env.Data); kv != nil {
		kv["ok"] = "true"
		return p.renderKVTable(kv, false)
	}

	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}

func (p *Printer) renderKVTable(kv map[string]string, isError bool) error {
	if isError {
		p.okLine(false)
	} else if v, ok := kv["ok"]; ok && v == "true" {
		p.okLine(true)
	}
	t := table.NewWriter()
	t.SetOutputMirror(p.Out)
	t.SetStyle(p.tableStyle())
	t.AppendHeader(table.Row{"字段", "值"})
	keys := sortedKeys(kv)
	for _, k := range keys {
		if v := kv[k]; v != "" {
			row := table.Row{k, v}
			if p.Color && k == "error" {
				row = table.Row{text.FgRed.Sprint(k), text.FgRed.Sprint(v)}
			}
			if p.Color && k == "hint" {
				row = table.Row{text.FgYellow.Sprint(k), text.FgYellow.Sprint(v)}
			}
			t.AppendRow(row)
		}
	}
	t.Render()
	return nil
}

// flattenData converts common map payloads to string key-values for table display.
func flattenData(data any) map[string]string {
	switch m := data.(type) {
	case map[string]string:
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out
	case map[string]any:
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = fmt.Sprint(v)
		}
		return out
	default:
		return nil
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		pi, pj := fieldOrder(keys[i]), fieldOrder(keys[j])
		if pi != pj {
			return pi < pj
		}
		return keys[i] < keys[j]
	})
	return keys
}

func fieldOrder(k string) int {
	switch k {
	case "ok":
		return 0
	case "error":
		return 1
	case "hint":
		return 2
	default:
		return 10
	}
}

func tabularRows(data any) ([]table.Row, bool) {
	list, ok := data.([]map[string]any)
	if !ok || len(list) == 0 {
		return nil, false
	}
	headers := make([]any, 0, len(list[0]))
	keys := make([]string, 0, len(list[0]))
	for k := range list[0] {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return nil, false
	}
	sort.Strings(keys)
	for _, k := range keys {
		headers = append(headers, k)
	}
	rows := []table.Row{headers}
	for _, item := range list {
		row := make(table.Row, len(keys))
		for i, k := range keys {
			row[i] = fmt.Sprint(item[k])
		}
		rows = append(rows, row)
	}
	return rows, true
}
