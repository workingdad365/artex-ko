package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/artex/guard"
	"github.com/Autumn-27/artex/intercept"
	"github.com/Autumn-27/norma/agentcore"
	"github.com/Autumn-27/norma/llm"
	"github.com/Autumn-27/norma/permission"
	"github.com/Autumn-27/norma/tool"
)

// Exercise the actual SDK event -> hook -> execution -> result path. No real
// model or command is used; the probe tool only returns a fixed string.
func TestCaptureApprovalLifecycle(t *testing.T) {
	dsn, _, err := db.DSN()
	if err != nil {
		t.Skip("no test database configured")
	}
	d, err := db.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	ic := intercept.New(d)
	priorTools, err := ic.GetEnabledTools()
	if err != nil {
		t.Fatal(err)
	}
	priorConfig := ic.GetJudgeConfig()
	t.Cleanup(func() { _ = ic.SetEnabledTools(priorTools); _ = ic.SetJudgeConfig(priorConfig) })
	if err := ic.SetEnabledTools([]string{"ApprovalAuditProbe"}); err != nil {
		t.Fatal(err)
	}
	if err := ic.SetJudgeConfig(intercept.JudgeConfig{Enabled: true, AskTimeoutSeconds: 1, AskTimeoutAction: "allow"}); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name, action, status, execution string
		manual, approve, toolError      bool
	}{
		{"model_fallback", "invalid", "allowed", "succeeded", false, false, false},
		{"model_allow", "allow", "allowed", "succeeded", false, false, false},
		{"model_deny", "deny", "denied", "not_executed", false, false, false},
		{"human_allow_tool_error", "ask", "allowed", "failed", true, true, true},
		{"human_deny", "ask", "denied", "not_executed", true, false, false},
		{"timeout_allow", "ask", "timeout", "succeeded", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ic.SetReviewer(func(context.Context, int64, string, intercept.ReviewInput) (intercept.Decision, error) {
				return intercept.Decision{Action: tc.action, Message: "probe review", ProfileID: 7}, nil
			})
			g := guard.NewWithInterceptor(ic)
			taskID := "approval-lifecycle-" + tc.name
			t.Cleanup(func() { _, _ = d.Exec(`DELETE FROM intercept_pending WHERE task_id=$1`, taskID) })
			ctx := intercept.WithTaskContext(t.Context(), taskID, "test-agent", func(a db.Activity) {
				if a.Kind != "intercept_request" || !tc.manual {
					return
				}
				var detail struct {
					ID int64 `json:"pending_id"`
				}
				if err := json.Unmarshal([]byte(a.Detail), &detail); err != nil {
					t.Error(err)
					return
				}
				if err := ic.Decide(detail.ID, tc.approve); err != nil {
					t.Error(err)
				}
			})
			turn, executions := 0, 0
			provider := captureUsageProvider{stream: func(_ context.Context, yield func(llm.StreamEvent, error) bool) {
				turn++
				events := []llm.StreamEvent{{Type: llm.SETextDelta, Text: "done"}, {Type: llm.SEMessageDelta, StopReason: "end_turn"}}
				if turn == 1 {
					events = []llm.StreamEvent{
						{Type: llm.SEToolUseStart, ToolID: "probe-call", ToolName: "ApprovalAuditProbe"},
						{Type: llm.SEToolInputJSON, Text: `{}`}, {Type: llm.SEMessageDelta, StopReason: "tool_use"},
					}
				}
				for _, event := range events {
					if !yield(event, nil) {
						return
					}
				}
			}}
			probe := tool.Build(tool.Spec{Name: "ApprovalAuditProbe", Schema: map[string]any{"type": "object"},
				Run: func(context.Context, json.RawMessage, *tool.ToolContext) (tool.Result, error) {
					executions++
					if tc.toolError {
						return tool.Errorf("probe failed"), nil
					}
					return tool.Text("probe succeeded"), nil
				},
			})
			_, _, err := captureRun(ctx, agentcore.Options{Provider: provider, Tools: []tool.CoreTool{probe}, Hooks: g.Hooks(),
				PermissionMode: permission.ModeBypass, WorkingDir: t.TempDir(), MaxTurns: 2}, "record this review", nil)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := d.ListTaskIntercepts(taskID)
			if err != nil || len(rows) != 1 {
				t.Fatalf("rows=%d err=%v", len(rows), err)
			}
			detail, err := d.GetInterceptDetail(rows[0].ID)
			if err != nil {
				t.Fatal(err)
			}
			a := detail.Audit
			initialAction := tc.action
			if tc.action == "invalid" {
				initialAction = "allow"
			}
			wantUserMessage := "record this review"
			if initialAction == "allow" {
				// Routine automatic allows keep decision/execution metadata but
				// intentionally omit the bulky replay context from the audit row.
				wantUserMessage = ""
			}
			if detail.Status != tc.status || detail.DecisionSource != "model" || a == nil || a.ExecutionStatus != tc.execution || a.ToolUseID != "probe-call" || a.UserMessage != wantUserMessage || a.InitialAction != initialAction || a.ModelFallback != (tc.action == "invalid") || a.ProfileID != 7 {
				t.Fatalf("unexpected review: row=%+v audit=%+v", detail.InterceptApprovalRow, a)
			}
			if initialAction == "allow" && (len(a.Context) != 0 || a.UserTruncated || a.ContextTruncated) {
				t.Fatalf("routine allow retained replay context: %+v", a)
			}
			wantCalls := 1
			if tc.execution == "not_executed" {
				wantCalls = 0
			}
			if executions != wantCalls {
				t.Fatalf("tool ran %d times, wanted %d", executions, wantCalls)
			}
		})
	}
}

// 표시용 번역이 다음 판정의 감사 문맥이나 SDK 대화 기록에 섞이지 않는지 검증한다.
func TestCaptureShellErrorPreservesModelAndAuditRecords(t *testing.T) {
	dsn, _, err := db.DSN()
	if err != nil {
		t.Skip("no test database configured")
	}
	d, err := db.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	t.Setenv("TMPDIR", t.TempDir())
	t.Setenv("AGENT_CORE_DISABLE_BACKGROUND_TASKS", "")
	t.Setenv("AGENT_CORE_DISABLE_INTERACTIVE_SHELL", "")
	ic := intercept.New(d)
	priorTools, err := ic.GetEnabledTools()
	if err != nil {
		t.Fatal(err)
	}
	priorConfig := ic.GetJudgeConfig()
	t.Cleanup(func() { _ = ic.SetEnabledTools(priorTools); _ = ic.SetJudgeConfig(priorConfig) })
	if err := ic.SetEnabledTools([]string{"shell_open"}); err != nil {
		t.Fatal(err)
	}
	if err := ic.SetJudgeConfig(intercept.JudgeConfig{Enabled: true}); err != nil {
		t.Fatal(err)
	}
	ic.SetReviewer(func(_ context.Context, _ int64, _ string, input intercept.ReviewInput) (intercept.Decision, error) {
		action := "allow"
		if strings.Contains(string(input.Arguments), "command") {
			action = "deny"
		}
		return intercept.Decision{Action: action, Message: "테스트 판정"}, nil
	})
	const taskID = "shell-display-audit"
	t.Cleanup(func() { _, _ = d.Exec(`DELETE FROM intercept_pending WHERE task_id=$1`, taskID) })
	ctx := intercept.WithTaskContext(t.Context(), taskID, "test-agent", nil)
	turn := 0
	provider := captureUsageProvider{stream: func(_ context.Context, yield func(llm.StreamEvent, error) bool) {
		turn++
		events := []llm.StreamEvent{{Type: llm.SETextDelta, Text: "완료"}, {Type: llm.SEMessageDelta, StopReason: "end_turn"}}
		if turn <= 2 {
			id, input := "shell-error", `{}`
			if turn == 2 {
				id, input = "shell-denied", `{"command":"unused"}`
			}
			events = []llm.StreamEvent{
				{Type: llm.SEToolUseStart, ToolID: id, ToolName: "shell_open"},
				{Type: llm.SEToolInputJSON, Text: input}, {Type: llm.SEMessageDelta, StopReason: "tool_use"},
			}
		}
		for _, event := range events {
			if !yield(event, nil) {
				return
			}
		}
	}}
	session := agentcore.NewSession(agentcore.Options{Provider: provider, Tools: []tool.CoreTool{tool.NewShellOpen()},
		Hooks: guard.NewWithInterceptor(ic).Hooks(), PermissionMode: permission.ModeBypass, WorkingDir: t.TempDir(), MaxTurns: 3})
	defer session.Close()
	results := map[string]db.Activity{}
	_, _, err = captureRunSession(ctx, session, "표시 검증", func(activity db.Activity) {
		if activity.Kind == "tool_result" {
			results[activity.ToolUseID] = activity
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	shown := results["shell-error"]
	if !shown.IsError || shown.Detail != "command는 필수입니다" || shown.Summary != shown.Detail {
		t.Fatalf("untranslated activity: %+v", shown)
	}
	if denied := results["shell-denied"]; !denied.IsError || !strings.HasPrefix(denied.Detail, "[ARTEX 플랫폼 제어]") || !strings.Contains(denied.Detail, "[모델] 테스트 판정") {
		t.Fatalf("untranslated guard activity: %+v", denied)
	}
	rawResults := map[string]string{}
	for _, entry := range approvalHistory(session.Messages()) {
		if entry.Kind == "tool_result" {
			rawResults[entry.ToolUseID] = entry.Text
		}
	}
	if rawResults["shell-error"] != "command 必填" || !strings.Contains(rawResults["shell-denied"], "[模型]") {
		t.Fatalf("model history changed: %+v", rawResults)
	}
	rows, err := d.ListTaskIntercepts(taskID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("audit rows=%d err=%v", len(rows), err)
	}
	for _, row := range rows {
		detail, err := d.GetInterceptDetail(row.ID)
		if err != nil || detail == nil || detail.Audit == nil {
			t.Fatalf("missing audit: %+v %v", detail, err)
		}
		audit := detail.Audit
		switch audit.ToolUseID {
		case "shell-error":
			if audit.ExecutionStatus != "failed" || audit.Output != rawResults["shell-error"] {
				t.Fatalf("raw execution result changed: %+v", audit)
			}
		case "shell-denied":
			found := false
			for _, entry := range audit.Context {
				if entry.Kind == "tool_result" && entry.ToolUseID == "shell-error" {
					found = true
					if entry.Text != rawResults["shell-error"] || !entry.IsError {
						t.Fatalf("raw audit context changed: %+v", entry)
					}
				}
			}
			if !found {
				t.Fatal("previous tool result missing from audit context")
			}
		default:
			t.Fatalf("unexpected audit: %+v", audit)
		}
	}
}
