package common

import (
	"context"

	"github.com/kuaimai-cli/kuaimai-cli/internal/client"
	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/core"
	"github.com/kuaimai-cli/kuaimai-cli/pkg/logger"
)

// RunGET executes a GET list/detail shortcut.
func RunGET(path string) error {
	return RunGETShortcut("", path)
}

// RunGETShortcut executes a GET list/detail shortcut against a configured shortcut base URL.
func RunGETShortcut(shortcut, path string) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	return NewRunner(f).ExecuteList(context.Background(), ListOptions{
		Method:  "GET",
		Path:    path,
		BaseURL: f.Config.ShortcutAPIURL(shortcut),
	})
}

// RunPOST executes a POST shortcut (JSON body when non-nil).
func RunPOST(path string, body map[string]any) error {
	return RunPOSTShortcut("", path, body)
}

// RunPOSTShortcut executes a JSON POST against a configured shortcut base URL.
func RunPOSTShortcut(shortcut, path string, body map[string]any) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	return NewRunner(f).ExecuteWrite(context.Background(), WriteOptions{
		Method:  "POST",
		Path:    path,
		Body:    body,
		BaseURL: f.Config.ShortcutAPIURL(shortcut),
	})
}

// RunPOSTJSON parses body JSON and POSTs to path.
func RunPOSTJSON(path, bodyJSON string) error {
	body, err := ParseBodyJSON(bodyJSON)
	if err != nil {
		return err
	}
	return RunPOST(path, body)
}

// RunPOSTForm parses body JSON and POSTs as application/x-www-form-urlencoded.
func RunPOSTForm(path, bodyJSON string) error {
	body, err := ParseBodyJSON(bodyJSON)
	if err != nil {
		return err
	}
	return RunPOSTFormMap(path, body)
}

// RunPOSTFormMap POSTs a form body map as application/x-www-form-urlencoded.
func RunPOSTFormMap(path string, body map[string]any) error {
	return RunPOSTFormMapShortcut("", path, body)
}

// RunPOSTFormMapShortcut POSTs a form body map against a configured shortcut base URL.
func RunPOSTFormMapShortcut(shortcut, path string, body map[string]any) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	return NewRunner(f).ExecuteWrite(context.Background(), WriteOptions{
		Method:      "POST",
		Path:        path,
		Body:        body,
		FormEncoded: true,
		BaseURL:     f.Config.ShortcutAPIURL(shortcut),
	})
}

// RunPOSTFormListMap POSTs a form list endpoint; supports --page-all via pageNo/pageSize.
func RunPOSTFormListMap(path string, body map[string]any) error {
	return RunPOSTFormListMapShortcut("", path, body)
}

// RunPOSTFormListMapShortcut POSTs a form list endpoint against a configured shortcut base URL.
func RunPOSTFormListMapShortcut(shortcut, path string, body map[string]any) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	return NewRunner(f).ExecuteWithBase(context.Background(), f.Config.ShortcutAPIURL(shortcut), func(ctx context.Context, c *client.Client) (any, error) {
		var data any
		var err error
		if core.Ctx.PageAll {
			logger.Info("page-all: 自动拉取全部分页（form pageNo/pageSize）")
			data, err = c.PostFormAllPages(ctx, path, body)
		} else {
			data, _, err = c.PostForm(ctx, path, client.MapToFormValues(body))
		}
		if err != nil {
			return nil, err
		}
		return NormalizeList(data), nil
	})
}
