package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
)

func writeCSV(out io.Writer, env Envelope) error {
	if !env.OK {
		return writeJSONEnvelope(out, env)
	}
	rows, ok := tabularRows(env.Data)
	if !ok {
		return writeJSONEnvelope(out, env)
	}
	w := csv.NewWriter(out)
	if len(rows) == 0 {
		w.Flush()
		return w.Error()
	}
	header := rowToStrings(rows[0])
	if err := w.Write(header); err != nil {
		return err
	}
	for i := 1; i < len(rows); i++ {
		if err := w.Write(rowToStrings(rows[i])); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func writeNDJSON(out io.Writer, env Envelope) error {
	if !env.OK {
		return writeJSONEnvelope(out, env)
	}
	items := ndjsonItems(env.Data)
	if len(items) == 0 {
		enc := json.NewEncoder(out)
		return enc.Encode(env)
	}
	enc := json.NewEncoder(out)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return err
		}
	}
	return nil
}

func writeJSONEnvelope(out io.Writer, env Envelope) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}

func rowToStrings(row table.Row) []string {
	out := make([]string, len(row))
	for i, cell := range row {
		out[i] = fmt.Sprint(cell)
	}
	return out
}

func ndjsonItems(data any) []map[string]any {
	switch v := data.(type) {
	case []map[string]any:
		return v
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case map[string]any:
		return []map[string]any{v}
	default:
		return nil
	}
}

// CSVHeaders returns sorted header keys from the first list row.
func CSVHeaders(list []map[string]any) []string {
	if len(list) == 0 {
		return nil
	}
	keys := make([]string, 0, len(list[0]))
	for k := range list[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
