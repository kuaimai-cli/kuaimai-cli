package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kuaimai/kuaimai-cli/pkg/util"
)

// Record appends a single audit line for business commands.
func Record(command, method, path, status string) error {
	dir := util.ConfigDir()
	if dir == "" {
		return nil
	}
	pathFile := filepath.Join(dir, "audit.log")
	f, err := os.OpenFile(pathFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
		time.Now().UTC().Format(time.RFC3339),
		sanitizeField(command),
		strings.ToUpper(method),
		sanitizeField(path),
		sanitizeField(status),
	)
	_, err = f.WriteString(line)
	return err
}

func sanitizeField(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
