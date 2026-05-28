package doctorcmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kuaimai/kuaimai-cli/internal/build"
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/kuaimai/kuaimai-cli/pkg/util"
	"github.com/spf13/cobra"
)

// Register adds doctor (install readiness check) command.
func Register(root *cobra.Command) {
	root.AddCommand(doctorCmd())
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "安装自检：配置、鉴权、PATH 与 Skill 提示",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}

			_, err = os.Stat(config.ConfigPath())
			configOK := err == nil

			loggedIn := f.Auth.IsLoggedIn()
			path, _ := exec.LookPath("kuaimai-cli")
			skillDir := filepath.Join(util.ConfigDir(), "..", ".agents", "skills", "kuaimai-item")
			// prefer ~/.agents/skills
			home, _ := os.UserHomeDir()
			if home != "" {
				skillDir = filepath.Join(home, ".agents", "skills", "kuaimai-item")
			}
			_, skillErr := os.Stat(filepath.Join(skillDir, "SKILL.md"))
			skillOK := skillErr == nil

			checks := []map[string]any{
				{"name": "config", "ok": configOK, "hint": hintConfig(configOK)},
				{"name": "auth", "ok": loggedIn, "hint": hintAuth(loggedIn)},
				{"name": "path", "ok": path != "", "hint": hintPath(path)},
				{"name": "skill_kuaimai_item", "ok": skillOK, "hint": hintSkill(skillOK)},
			}
			allOK := configOK && loggedIn && path != "" && skillOK

			return f.Printer().Success(map[string]any{
				"version": build.Version,
				"ready":   allOK,
				"checks":  checks,
				"next":    nextSteps(configOK, loggedIn, skillOK),
			})
		},
	}
}

func hintConfig(ok bool) string {
	if ok {
		return "config.yaml 已存在"
	}
	return "执行 kuaimai-cli config init"
}

func hintAuth(ok bool) string {
	if ok {
		return "已登录"
	}
	return "执行 kuaimai-cli auth login <accessToken>"
}

func hintPath(p string) string {
	if p != "" {
		return p
	}
	return "将 kuaimai-cli 加入 PATH，或使用 npx @kuaimai/cli"
}

func hintSkill(ok bool) string {
	if ok {
		return "kuaimai-item Skill 已安装"
	}
	return "执行 kuaimai-cli skill install-all --from ./skills"
}

func nextSteps(configOK, loggedIn, skillOK bool) []string {
	var steps []string
	if !configOK {
		steps = append(steps, "kuaimai-cli config init")
	}
	if !loggedIn {
		steps = append(steps, "kuaimai-cli auth login <accessToken>")
	}
	if loggedIn {
		steps = append(steps, "kuaimai-cli auth check")
	}
	if !skillOK {
		steps = append(steps, "kuaimai-cli skill install-all --from ./skills")
	}
	if len(steps) == 0 {
		steps = append(steps, "环境就绪，可使用 item +list / item update-title 等命令")
	}
	return steps
}
