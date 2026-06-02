package core

import (
	"github.com/kuaimai-cli/kuaimai-cli/internal/output"
	"github.com/kuaimai-cli/kuaimai-cli/internal/pagination"
)

// Global holds process-wide CLI state set during root command init.
type Global struct {
	Verbose     bool
	DryRun      bool
	PageAll     bool
	PageLimit   int
	PageConfirm string
	Output      output.Format
	NoColor     bool
}

// Ctx is the active global context for the current invocation.
var Ctx Global

// PaginationSettings returns page-all safety settings from global flags.
func PaginationSettings() pagination.Settings {
	return pagination.Settings{
		MaxPages:    pagination.DefaultMaxPages,
		RecordLimit: Ctx.PageLimit,
		ConfirmMode: Ctx.PageConfirm,
	}
}
