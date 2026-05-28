package core

import (
	"github.com/kuaimai/kuaimai-cli/internal/output"
)

// Global holds process-wide CLI state set during root command init.
type Global struct {
	Verbose  bool
	DryRun   bool
	PageAll  bool
	Output   output.Format
	NoColor bool
}

// Ctx is the active global context for the current invocation.
var Ctx Global
