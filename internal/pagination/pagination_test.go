package pagination

import (
	"bytes"
	"strings"
	"testing"
)

func TestCollectPages_complete(t *testing.T) {
 calls := 0
 res, err := CollectPages(Settings{ConfirmMode: ConfirmYes}, 1, 50, func(page, size int) ([]map[string]any, bool, int, error) {
  calls++
  if page == 1 {
   items := make([]map[string]any, 50)
   for i := range items {
    items[i] = map[string]any{"id": i}
   }
   return items, true, 80, nil
  }
  items := make([]map[string]any, 30)
  for i := range items {
   items[i] = map[string]any{"id": 50 + i}
  }
  return items, false, 80, nil
 })
 if err != nil {
  t.Fatal(err)
 }
 if res.Truncated || res.Fetched != 80 || len(res.Items) != 80 {
  t.Fatalf("got truncated=%v fetched=%d len=%d", res.Truncated, res.Fetched, len(res.Items))
 }
 if calls != 2 {
  t.Fatalf("expected 2 page calls, got %d", calls)
 }
}

func TestCollectPages_recordLimit(t *testing.T) {
 res, err := CollectPages(Settings{RecordLimit: 60, ConfirmMode: ConfirmYes}, 1, 50, func(page, size int) ([]map[string]any, bool, int, error) {
  items := make([]map[string]any, 50)
  for i := range items {
   items[i] = map[string]any{"page": page}
  }
  return items, true, 500, nil
 })
 if err != nil {
  t.Fatal(err)
 }
 if !res.Truncated || res.Reason != StopRecordLimit || res.Fetched != 60 {
  t.Fatalf("got %+v", res)
 }
}

func TestCollectPages_userDeclined(t *testing.T) {
 var stdin bytes.Buffer
 stdin.WriteString("n\n")
 var stderr bytes.Buffer
 page := 0
 res, err := CollectPages(Settings{
  ConfirmMode: ConfirmPrompt,
  IsInteractive: func() bool { return true },
  Stdin:       &stdin,
  Stderr:      &stderr,
 }, 1, 50, func(p, size int) ([]map[string]any, bool, int, error) {
  page++
  items := make([]map[string]any, 50)
  return items, true, 5000, nil
 })
 if err != nil {
  t.Fatal(err)
 }
 if !res.Truncated || res.Reason != StopUserDeclined || res.Fetched != 50 {
  t.Fatalf("got %+v", res)
 }
 if page != 1 {
  t.Fatalf("expected stop after first page, got %d calls", page)
 }
 if !strings.Contains(stderr.String(), "是否继续") {
  t.Fatalf("expected prompt on stderr: %q", stderr.String())
 }
}

func TestCollectPages_nonInteractiveThreshold(t *testing.T) {
 res, err := CollectPages(Settings{
  ConfirmMode: ConfirmPrompt,
  IsInteractive: func() bool { return false },
 }, 1, 50, func(page, size int) ([]map[string]any, bool, int, error) {
  items := make([]map[string]any, 50)
  return items, true, 5000, nil
 })
 if err != nil {
  t.Fatal(err)
 }
 if res.Reason != StopNonInteractive || res.Fetched != 50 {
  t.Fatalf("got %+v", res)
 }
}

func TestCollectPages_confirmYesBypassesThreshold(t *testing.T) {
 pages := 0
 res, err := CollectPages(Settings{ConfirmMode: ConfirmYes}, 1, 50, func(page, size int) ([]map[string]any, bool, int, error) {
  pages++
  if pages > 12 {
   return nil, false, 600, nil
  }
  items := make([]map[string]any, 50)
  return items, true, 600, nil
 })
 if err != nil {
  t.Fatal(err)
 }
 if res.Fetched != 600 {
  t.Fatalf("expected 600 fetched, got %d reason=%s", res.Fetched, res.Reason)
 }
}

func TestChunkCollector_releasesChunks(t *testing.T) {
 c := newChunkCollector(2)
 c.add([]map[string]any{{"a": 1}, {"a": 2}})
 c.add([]map[string]any{{"a": 3}})
 out := c.result()
 if len(out) != 3 {
  t.Fatalf("len=%d", len(out))
 }
}
