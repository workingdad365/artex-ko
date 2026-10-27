package agent

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"unicode"

	"github.com/Autumn-27/norma/tool"
)

// SDK가 실제 반환하는 오류를 검증해 의존성 갱신 시 번역 누락을 발견한다.
func TestSDKShellErrorsKoreanDisplay(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	t.Setenv("AGENT_CORE_DISABLE_INTERACTIVE_SHELL", "")
	m, err := tool.NewManager("shell-localization")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Cleanup)
	tc := &tool.ToolContext{Tasks: m}
	check := func(t *testing.T, core tool.CoreTool, input string, tc *tool.ToolContext) {
		t.Helper()
		result, err := core.Call(t.Context(), json.RawMessage(input), tc)
		if err != nil || !result.IsError {
			t.Fatalf("expected SDK validation error: %+v %v", result, err)
		}
		raw := result.Flatten()
		shown := toolResultDisplay(core.Name(), raw, result.IsError)
		if shown == raw || strings.ContainsFunc(shown, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Fatalf("untranslated SDK error: %q -> %q", raw, shown)
		}
		if result.Flatten() != raw {
			t.Fatal("SDK result mutated")
		}
	}
	for _, core := range tool.ShellSessionTools() {
		t.Run(core.Name()+"/missing_manager", func(t *testing.T) { check(t, core, `{}`, nil) })
		t.Run(core.Name()+"/disabled", func(t *testing.T) {
			t.Setenv("AGENT_CORE_DISABLE_INTERACTIVE_SHELL", "1")
			check(t, core, `{}`, tc)
		})
	}
	for _, input := range []string{
		`{}`,
		`{"command":"unused","mode":"invalid"}`,
		`{"command":"unused","mode":"monitor"}`,
		`{"command":"unused","mode":"monitor","watch":["["]}`,
		`{"command":"unused","prompt_regex":"["}`,
	} {
		t.Run("shell_open/"+input, func(t *testing.T) { check(t, tool.NewShellOpen(), input, tc) })
	}
	for _, core := range []tool.CoreTool{tool.NewShellSend(), tool.NewShellRead(), tool.NewShellClose()} {
		t.Run(core.Name()+"/missing_session", func(t *testing.T) { check(t, core, `{"session_id":"missing"}`, tc) })
	}
	if runtime.GOOS == "windows" {
		return
	}
	result, err := tool.NewShellOpen().Call(t.Context(), json.RawMessage(`{"command":"printf 'command 必填\\n'","mode":"interactive","startup_grace_ms":1,"quiet_ms":100}`), tc)
	if err != nil || result.IsError {
		t.Fatalf("shell_open failed: %+v %v", result, err)
	}
	var opened struct {
		ID     string `json:"session_id"`
		Output string `json:"output"`
	}
	if err := json.Unmarshal([]byte(result.Flatten()), &opened); err != nil || !strings.Contains(opened.Output, "command 必填") {
		t.Fatalf("missing original shell output: %s %v", result.Flatten(), err)
	}
	if got := toolResultDisplay("shell_open", result.Flatten(), result.IsError); got != result.Flatten() {
		t.Fatalf("shell output changed: %s", got)
	}
	input, _ := json.Marshal(map[string]string{"session_id": opened.ID, "text": "unused"})
	check(t, tool.NewShellSend(), string(input), tc)
}

func TestToolResultDisplayPreservesEvidence(t *testing.T) {
	const evidence = "command 必填"
	for _, toolName := range []string{"Bash", "custom_tool", "shell_open"} {
		for _, isError := range []bool{false, true} {
			for _, output := range []string{
				`{"output":"command 必填","screen":"会话不存在: 原文"}`,
				"HTTP 400: command 必填",
			} {
				if got := toolResultDisplay(toolName, output, isError); got != output {
					t.Fatalf("evidence changed for %s: %q", toolName, got)
				}
			}
			if toolName != "shell_open" || !isError {
				if got := toolResultDisplay(toolName, evidence, isError); got != evidence {
					t.Fatalf("non-SDK error changed for %s: %q", toolName, got)
				}
			}
		}
	}
	for _, tc := range []struct{ raw, want string }{
		{"会话不存在: 用户会话", "세션이 없습니다: 用户会话"},
		{"未知按键: 未知按键", "알 수 없는 키: 未知按键"},
		{"watch 正则无效 [原文: parse error", "watch 정규식이 올바르지 않습니다 [原文: parse error"},
		{"写入失败: 原文", "쓰기 실패: 原文"},
		{"开启会话失败: 原文", "세션 열기 실패: 原文"},
		{"开启会话失败: 交互式 shell(PTY) 在当前平台不支持", "세션 열기 실패: 현재 플랫폼은 대화형 shell(PTY)을 지원하지 않습니다"},
	} {
		if got := toolResultDisplay("shell_open", tc.raw, true); got != tc.want {
			t.Fatalf("error framing or original argument changed: %q, want %q", got, tc.want)
		}
	}
}
