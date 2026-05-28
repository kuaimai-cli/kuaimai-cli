package client

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kuaimai/kuaimai-cli/internal/auth"
)

// BusinessError is returned when HTTP succeeds but the API body reports failure (e.g. result=901).
type BusinessError struct {
	Result       int64
	Message      string
	Hint         string
	TokenPreview string
}

func (e *BusinessError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("业务错误 result=%d", e.Result)
}

// ParseBusinessError inspects a parsed JSON body for 快麦 CommonResponse conventions (result=1 为成功).
func ParseBusinessError(body any) *BusinessError {
	m, ok := body.(map[string]any)
	if !ok {
		return nil
	}
	code, ok := resultCode(m["result"])
	if !ok {
		return nil
	}
	if code == 1 {
		return nil
	}
	msg, _ := m["message"].(string)
	if msg == "" {
		msg = fmt.Sprintf("业务错误 result=%d", code)
	}
	return &BusinessError{
		Result:  code,
		Message: msg,
		Hint:    hintForBusinessResult(code),
	}
}

func resultCode(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

func hintForBusinessResult(code int64) string {
	switch code {
	case 901:
		return "会话已失效。请执行 kuaimai-cli auth login <accessToken> 重新登录，并确认 token 与 config 中 api.url 为同一环境（如 scm1.superboss.cc）。使用 kuaimai-cli auth status 查看是否已登录。"
	default:
		return "请检查请求参数或联系平台管理员；使用 --verbose 查看请求详情"
	}
}

// EnrichBusinessError attaches auth context for clearer session failure messages.
func EnrichBusinessError(err *BusinessError, tokenPreview string) *BusinessError {
	if err == nil {
		return nil
	}
	err.TokenPreview = tokenPreview
	if err.Result == 901 && tokenPreview != "" {
		err.Hint = fmt.Sprintf(
			"请求头 %s 已携带（预览: %s），但服务端返回会话异常（result=901）。请重新执行 kuaimai-cli auth login <accessToken>，并确认 token 与 config 中 api.url 为同一环境；可用 kuaimai-cli auth status 核对。使用 --verbose 查看完整请求。",
			auth.HeaderAccessToken,
			tokenPreview,
		)
	}
	return err
}
