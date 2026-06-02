package upgrade

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const npmPackage = "@kuaimai-cli/cli@latest"

// Apply upgrades the CLI via npm global install (same path as npx install wizard).
func Apply() error {
	npm, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("未找到 npm，请手动运行: npx %s install", npmPackage)
	}
	cmd := exec.Command(npm, "install", "-g", npmPackage)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm 升级失败: %w", err)
	}
	return nil
}

// ApplyHint returns post-upgrade guidance for stderr.
func ApplyHint(latest string) string {
	latest = strings.TrimSpace(latest)
	if latest == "" {
		latest = "latest"
	}
	return fmt.Sprintf(
		"已触发 npm 全局升级至 %s。请重新打开终端并执行 kuaimai-cli --version 确认；若版本未变，运行: npx @kuaimai-cli/cli@latest install",
		strings.TrimPrefix(latest, "v"),
	)
}
