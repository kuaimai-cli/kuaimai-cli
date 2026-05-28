package cmdutil

import (
	"github.com/kuaimai/kuaimai-cli/internal/auth"
	"github.com/kuaimai/kuaimai-cli/internal/client"
	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/kuaimai/kuaimai-cli/internal/core"
	"github.com/kuaimai/kuaimai-cli/internal/output"
)

// Factory wires config, auth, HTTP client, and output for commands.
type Factory struct {
	Config *config.Manager
	Auth   *auth.Store
}

// NewFactory builds dependencies for one CLI invocation.
func NewFactory() (*Factory, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}
	store, err := auth.NewStore()
	if err != nil {
		return nil, err
	}
	return &Factory{Config: cfg, Auth: store}, nil
}

// HTTPClient returns an API client with global dry-run flag and request headers.
func (f *Factory) HTTPClient() (*client.Client, error) {
	return client.New(f.Config, f.Auth, core.Ctx.DryRun)
}

// Printer returns stdout printer with global output format and color settings.
func (f *Factory) Printer() *output.Printer {
	p := output.NewPrinter(core.Ctx.Output)
	color := f.Config.CLIColorEnabled() && !core.Ctx.NoColor
	p.SetColor(color)
	return p
}

// RequireAuth returns a friendly error if not logged in.
func (f *Factory) RequireAuth() error {
	if !f.Auth.IsLoggedIn() {
		return ErrNotLoggedIn
	}
	return nil
}

// ErrNotLoggedIn is returned when business commands run without auth.
var ErrNotLoggedIn = &AuthRequiredError{}

// AuthRequiredError indicates the user must log in first.
type AuthRequiredError struct{}

func (e *AuthRequiredError) Error() string {
	return "未登录"
}

func (e *AuthRequiredError) Hint() string {
	return "请先执行 kuaimai-cli auth login <accessToken> 完成登录"
}
