package cli_e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type envelope struct {
	OK    bool           `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error string         `json:"error"`
	Hint  string         `json:"hint"`
}

func TestSmokeConfigAuthDryRun(t *testing.T) {
	bin := buildCLI(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	runOK(t, bin, "config", "init")
	out := runOK(t, bin, "config", "get", "api.url", "--output", "json")
	var env envelope
	mustJSON(t, out, &env)
	if !env.OK {
		t.Fatalf("config get failed: %s", env.Error)
	}

	// init again should not overwrite
	runOK(t, bin, "config", "init")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/item/stock/queryCount" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("accessToken"); got != "test-token-12345678" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":1,"data":{"total":3}}`))
	}))
	defer srv.Close()

	tokenFile := filepath.Join(home, ".kuaimai-cli", "tokens.json")
	t.Setenv("KUAIMAI_CLI_TOKEN_FILE", tokenFile)

	runOK(t, bin, "config", "set", "api.url", srv.URL)
	runOK(t, bin, "auth", "login", "test-token-12345678")

	checkOut := runOK(t, bin, "auth", "check", "--output", "json")
	mustJSON(t, checkOut, &env)
	if !env.OK {
		t.Fatalf("auth check: %s %s", env.Error, env.Hint)
	}

	statusOut := runOK(t, bin, "auth", "status", "--output", "json")
	mustJSON(t, statusOut, &env)
	if !env.OK {
		t.Fatalf("auth status failed")
	}

	dryOut := runOK(t, bin, "item", "save", "--body", `{"sysItemId":1,"title":"x"}`, "--dry-run", "--output", "json")
	mustJSON(t, dryOut, &env)
	if !env.OK {
		t.Fatalf("dry-run save: %s", env.Error)
	}
	var dryData map[string]any
	if err := json.Unmarshal(env.Data, &dryData); err != nil || dryData["dry_run"] != true {
		t.Fatalf("expected dry_run=true in data: %s", env.Data)
	}

	schemaOut := runOK(t, bin, "schema", "--output", "json")
	mustJSON(t, schemaOut, &env)
	if !env.OK {
		t.Fatalf("schema failed")
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "kuaimai-cli")
	cmd := exec.Command("go", "build", "-mod=vendor", "-o", out, ".")
	cmd.Dir = projectRoot(t)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, b)
	}
	return out
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// tests/cli_e2e -> repo root
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func runOK(t *testing.T, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if v := os.Getenv("KUAIMAI_CLI_TOKEN_FILE"); v != "" {
		cmd.Env = append(cmd.Env, "KUAIMAI_CLI_TOKEN_FILE="+v)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v: stdout=%s stderr=%s err=%v", bin, args, stdout.String(), stderr.String(), err)
	}
	return stdout.String()
}

func mustJSON(t *testing.T, raw string, env *envelope) {
	t.Helper()
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), env); err != nil {
		t.Fatalf("json: %v raw=%s", err, raw)
	}
}
