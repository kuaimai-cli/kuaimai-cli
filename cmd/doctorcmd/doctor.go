package doctorcmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kuaimai-cli/kuaimai-cli/internal/auth"
	"github.com/kuaimai-cli/kuaimai-cli/internal/build"
	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/config"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/kuaimai-cli/kuaimai-cli/internal/skill"
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
			skillOK, _ := skill.IsInstalled("kuaimai-erp-item")
			refOK, _ := skill.HasReferences("kuaimai-erp-item")
			scmSkillOK, _ := skill.IsInstalled("kuaimai-scm-item")
			scmRefOK, _ := skill.HasReferences("kuaimai-scm-item")
			skillRoots, _ := skill.DefaultRootStatuses()
			skillReady := skillOK && refOK && scmSkillOK && scmRefOK && allSkillRootsOK(skillRoots)
			_, regVersion, regCount, regOK := registry.CacheInfo()

			checks := []map[string]any{
				{"name": "config", "ok": configOK, "hint": hintConfig(configOK)},
				{"name": "auth", "ok": loggedIn, "hint": hintAuth(loggedIn)},
				{"name": "path", "ok": path != "", "hint": hintPath(path)},
				{"name": "registry", "ok": regOK, "hint": hintRegistry(regOK, regVersion, regCount)},
				{"name": "skill_kuaimai_erp_item", "ok": skillOK && refOK, "hint": hintSkill(skillOK, refOK)},
				{"name": "skill_kuaimai_scm_item", "ok": scmSkillOK && scmRefOK, "hint": hintScmSkill(scmSkillOK, scmRefOK)},
			}
			allOK := configOK && loggedIn && path != "" && regOK && skillReady

			return f.Printer().Success(map[string]any{
				"version":     build.Version,
				"ready":       allOK,
				"checks":      checks,
				"skill_roots": skillRoots,
				"next":        nextSteps(configOK, loggedIn, regOK, skillReady),
			})
		},
	}
}

func allSkillRootsOK(statuses []skill.RootStatus) bool {
	if len(statuses) == 0 {
		return false
	}
	for _, st := range statuses {
		if !st.OK {
			return false
		}
	}
	return true
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
	return auth.LoginHint
}

func hintRegistry(ok bool, version string, count int) string {
	if ok {
		return fmt.Sprintf("registry 已同步（version=%s, apis=%d）", version, count)
	}
	return "执行 kuaimai-cli registry sync"
}

func hintPath(p string) string {
	if p != "" {
		return p
	}
	return "将 kuaimai-cli 加入 PATH，或使用 npx @kuaimai-cli/cli"
}

func hintSkill(installed, hasRefs bool) string {
	if installed && hasRefs {
		return "kuaimai-erp-item Skill 已安装（含 references/）"
	}
	if installed && !hasRefs {
		return "kuaimai-erp-item 缺少 references/，请执行 kuaimai-cli skill install 重装"
	}
	return "执行 kuaimai-cli skill install"
}

func hintScmSkill(installed, hasRefs bool) string {
	if installed && hasRefs {
		return "kuaimai-scm-item Skill 已安装（含 references/）"
	}
	if installed && !hasRefs {
		return "kuaimai-scm-item 缺少 references/，请执行 kuaimai-cli skill install kuaimai-scm-item"
	}
	return "执行 kuaimai-cli skill install kuaimai-scm-item"
}

func nextSteps(configOK, loggedIn, regOK, skillOK bool) []string {
	var steps []string
	if !configOK {
		steps = append(steps, "kuaimai-cli config init")
	}
	if !regOK {
		steps = append(steps, "kuaimai-cli registry sync")
	}
	if !loggedIn {
		steps = append(steps, auth.LoginHint)
	}
	if loggedIn {
		steps = append(steps, "kuaimai-cli auth check")
	}
	if !skillOK {
		steps = append(steps, "kuaimai-cli skill install")
	}
	if len(steps) == 0 {
		steps = append(steps, "环境就绪，可使用 erp-item +list / web call <apiId> 等命令")
	}
	return steps
}
