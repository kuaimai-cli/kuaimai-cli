package pagination

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Thresholds align with skills/kuaimai-item (Agent + CLI dual protection).
const (
	DefaultMaxPages        = 1000
	PromptRecordThreshold  = 500
	EstimateTotalThreshold = 1000
	ChunkSize              = 500

	ConfirmPrompt = "prompt"
	ConfirmYes    = "yes"
	ConfirmNo     = "no"
)

// StopReason describes why pagination stopped early.
type StopReason string

const (
	StopComplete         StopReason = "complete"
	StopRecordLimit      StopReason = "record_limit"
	StopMaxPages         StopReason = "max_pages"
	StopUserDeclined     StopReason = "user_declined"
	StopThresholdNo      StopReason = "threshold_no_confirm"
	StopNonInteractive   StopReason = "threshold_non_interactive"
)

// Settings controls page-all safety behavior.
type Settings struct {
	MaxPages      int
	RecordLimit   int
	ConfirmMode   string
	IsInteractive func() bool
	Stdin         io.Reader
	Stderr        io.Writer
}

// Result is the outcome of a paginated fetch.
type Result struct {
	Items     []map[string]any
	Truncated bool
	Reason    StopReason
	Fetched   int
	Notice    string
}

// PageFetch loads one page; estimatedTotal is from API total field when present.
type PageFetch func(pageNo, pageSize int) (items []map[string]any, hasMore bool, estimatedTotal int, err error)

// CollectPages runs the page loop with threshold prompts and chunked merging.
func CollectPages(settings Settings, startPage, pageSize int, fetch PageFetch) (Result, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	maxPages := settings.MaxPages
	if maxPages <= 0 {
		maxPages = DefaultMaxPages
	}
	mode := normalizeConfirmMode(settings.ConfirmMode)

	collector := newChunkCollector(ChunkSize)
	fetched := 0
	nextPromptAt := PromptRecordThreshold
	page := startPage
	endPage := startPage + maxPages

	for page < endPage {
		items, hasMore, estimatedTotal, err := fetch(page, pageSize)
		if err != nil {
			return Result{}, err
		}

		if settings.RecordLimit > 0 && fetched+len(items) > settings.RecordLimit {
			items = items[:settings.RecordLimit-fetched]
		}
		collector.add(items)
		fetched += len(items)

		if settings.RecordLimit > 0 && fetched >= settings.RecordLimit {
			return finish(collector, true, StopRecordLimit, fetched,
				fmt.Sprintf("page-all: 已达 --page-limit %d 条，停止翻页", settings.RecordLimit)), nil
		}

		if len(items) == 0 || !hasMore {
			return finish(collector, false, StopComplete, fetched, ""), nil
		}

		if shouldPrompt(fetched, estimatedTotal, nextPromptAt) {
			cont, notice, reason := resolveContinue(settings, mode, fetched, estimatedTotal)
			if !cont {
				return finish(collector, true, reason, fetched, notice), nil
			}
			if notice != "" {
				writeNotice(settings, notice)
			}
			nextPromptAt = nextPromptThreshold(fetched)
		}

		page++
	}

	return finish(collector, true, StopMaxPages, fetched,
		fmt.Sprintf("page-all: 已达最大翻页数 %d 页，停止翻页", maxPages)), nil
}

func normalizeConfirmMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ConfirmYes, ConfirmNo:
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return ConfirmPrompt
	}
}

func shouldPrompt(fetched, estimatedTotal, nextPromptAt int) bool {
	if fetched >= nextPromptAt {
		return true
	}
	if estimatedTotal > EstimateTotalThreshold {
		return true
	}
	return false
}

func nextPromptThreshold(fetched int) int {
	return ((fetched / PromptRecordThreshold) + 1) * PromptRecordThreshold
}

func resolveContinue(settings Settings, mode string, fetched, estimatedTotal int) (continueFetch bool, notice string, reason StopReason) {
	switch mode {
	case ConfirmYes:
		return true, "", ""
	case ConfirmNo:
		return false,
			fmt.Sprintf("page-all: 已拉取 %d 条，--page-confirm no 在阈值处停止", fetched),
			StopThresholdNo
	}

	if interactive(settings) {
		if askContinue(settings, fetched, estimatedTotal) {
			return true, "", ""
		}
		return false,
			fmt.Sprintf("page-all: 用户取消续查，返回已拉取的 %d 条数据", fetched),
			StopUserDeclined
	}

	msg := fmt.Sprintf("page-all: 已拉取 %d 条", fetched)
	if estimatedTotal > EstimateTotalThreshold {
		msg += fmt.Sprintf("，接口预估共 %d 条", estimatedTotal)
	}
	msg += "。非交互环境已达阈值并停止；使用 --page-confirm yes 继续，或 --page-limit 限制条数"
	return false, msg, StopNonInteractive
}

func interactive(settings Settings) bool {
	if settings.IsInteractive != nil {
		return settings.IsInteractive()
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func askContinue(settings Settings, fetched, estimatedTotal int) bool {
	out := settings.Stderr
	if out == nil {
		out = os.Stderr
	}
	msg := fmt.Sprintf("page-all: 当前已拉取 %d 条数据", fetched)
	if estimatedTotal > 0 {
		msg += fmt.Sprintf("，接口预估共 %d 条", estimatedTotal)
	}
	msg += "，数据量较大，是否继续翻页查询？[y/N]: "
	fmt.Fprint(out, msg)

	in := settings.Stdin
	if in == nil {
		in = os.Stdin
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

func writeNotice(settings Settings, notice string) {
	if notice == "" {
		return
	}
	out := settings.Stderr
	if out == nil {
		out = os.Stderr
	}
	fmt.Fprintln(out, notice)
}

func finish(collector *chunkCollector, truncated bool, reason StopReason, fetched int, notice string) Result {
	return Result{
		Items:     collector.result(),
		Truncated: truncated,
		Reason:    reason,
		Fetched:   fetched,
		Notice:    notice,
	}
}
