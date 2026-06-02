package upgrade

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const defaultCheckInterval = 24 * time.Hour

// ShouldSkipUpdateCheck reports whether background update notification should be skipped.
func ShouldSkipUpdateCheck(cmd *cobra.Command) bool {
	if strings.TrimSpace(os.Getenv("KUAIMAI_CLI_SKIP_UPDATE_CHECK")) != "" {
		return true
	}
	if cmd != nil && cmd.Root() != nil && cmd.Root().Flags().Changed("version") {
		return true
	}
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "upgrade", "doctor", "completion", "help", "version":
			return true
		}
	}
	return false
}

// MaybeNotify prints a one-line stderr hint when a newer release is available.
// Uses a 24h cache in ~/.kuaimai-cli/version-check.json.
func MaybeNotify(cmd *cobra.Command) {
	if ShouldSkipUpdateCheck(cmd) {
		return
	}
	res, err := checkWithCache("")
	if err != nil || res == nil || !res.UpdateAvail {
		return
	}
	latest := strings.TrimPrefix(res.Latest, "v")
	fmt.Fprintf(os.Stderr,
		"\n[kuaimai-cli] 发现新版本 %s（当前 %s）。运行: npx @kuaimai-cli/cli@latest install  或: kuaimai-cli upgrade\n",
		latest, res.Current,
	)
}

func checkWithCache(repo string) (*CheckResult, error) {
	st, _ := loadVersionCheckState()
	if st.LastResult != nil && time.Since(st.LastCheckAt) < defaultCheckInterval {
		return st.LastResult, nil
	}
	res, err := CheckLatest(repo)
	if err != nil {
		return nil, err
	}
	st.LastCheckAt = time.Now()
	st.LastResult = res
	_ = saveVersionCheckState(st)
	return res, nil
}
