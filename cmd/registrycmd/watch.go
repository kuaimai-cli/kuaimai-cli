package registrycmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/internal/config"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/spf13/cobra"
)

func watchCmd() *cobra.Command {
	var interval int
	var source string
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "监听远端 registry 变化并自动同步（开发/运维）",
		Long: `定时拉取 registry.source，感知 version/ETag 变化后写入本地缓存。
适合本地观察 api-onboard 发布后的接口更新。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(source, interval)
		},
	}
	cmd.Flags().IntVar(&interval, "interval", 30, "轮询间隔（秒）")
	cmd.Flags().StringVar(&source, "source", "", "registry.json URL（默认 config registry.source）")
	return cmd
}

func runWatch(sourceFlag string, intervalSec int) error {
	if intervalSec < 5 {
		return fmt.Errorf("interval 不能小于 5 秒")
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}
	source := sourceFlag
	if source == "" {
		source = cfg.RegistrySource()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stderr, "[kuaimai-cli] registry watch 已启动: %s  间隔 %ds  (Ctrl+C 停止)\n",
		source, intervalSec)

	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	runOnce := func() (*registry.SyncResult, error) {
		return registry.SyncIfNeeded(ctx, source)
	}

	if res, err := runOnce(); err != nil {
		fmt.Fprintf(os.Stderr, "[kuaimai-cli] 首次同步失败: %v\n", err)
	} else {
		printWatchResult(res, true)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "[kuaimai-cli] registry watch 已停止")
			return nil
		case <-ticker.C:
			res, err := runOnce()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[kuaimai-cli] 同步失败: %v\n", err)
				continue
			}
			if res.Changed {
				printWatchResult(res, false)
			}
		}
	}
}

func printWatchResult(res *registry.SyncResult, initial bool) {
	if res == nil {
		return
	}
	prefix := "同步"
	if initial {
		prefix = "当前"
	}
	if res.NotModified || !res.Changed {
		if initial {
			fmt.Fprintf(os.Stderr, "[kuaimai-cli] %s registry: version=%s apis=%d（已是最新）\n",
				prefix, res.Version, res.APICount)
		}
		return
	}
	fmt.Fprintf(os.Stderr, "[kuaimai-cli] %s registry 已更新: version=%s apis=%d path=%s\n",
		prefix, res.Version, res.APICount, res.Path)
	fmt.Fprintln(os.Stderr, "[kuaimai-cli] 提示: 执行 kuaimai-cli schema / web call 等命令将自动加载新版本")
}
