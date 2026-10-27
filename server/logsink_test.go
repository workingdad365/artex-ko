package server

import "testing"

func TestParseLogLocalizedLevels(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		level string
		tag   string
	}{
		{
			name:  "planner_completed",
			text:  "[planner] 작업 2 계획 수립 완료",
			level: "info",
			tag:   "planner",
		},
		{
			name:  "worker_state_write_failed",
			text:  "[worker work#3] 작업 2 의도 #34 최종 상태 blocked 저장 실패: sql: database is closed",
			level: "error",
			tag:   "worker work#3",
		},
		{
			name:  "activity_dropped",
			text:  "[activity] 작업 2 활동 기록 유실(해당 작업 누적 1건) worker=work#3",
			level: "error",
			tag:   "activity",
		},
		{
			name:  "enrichment_queue_full",
			text:  "[enrich] 대기열이 가득 차 작업 폐기 kind=1 id=2",
			level: "error",
			tag:   "enrich",
		},
		{
			name:  "auth_rejected",
			text:  "[auth] 초기 설정 거부: 설정 토큰 불일치 (client=127.0.0.1)",
			level: "error",
			tag:   "auth",
		},
		{
			name:  "planner_error",
			text:  "[planner] 작업 2 계획 수립 오류: context canceled",
			level: "error",
			tag:   "planner",
		},
		{
			name:  "search_disabled",
			text:  "[traffic] 전문 인덱스를 사용할 수 없어 본문 검색 비활성화(메타데이터 검색에는 영향 없음)",
			level: "warn",
			tag:   "traffic",
		},
		{
			name:  "update_skipped",
			text:  "[update] 시작 시 업데이트 처리 건너뜀: unsupported platform",
			level: "warn",
			tag:   "update",
		},
		{
			name:  "worker_stopped",
			text:  "[worker work#3] 작업 2 의도 #34 중지됨, 취소 정리 대기 중",
			level: "warn",
			tag:   "worker work#3",
		},
		{
			name:  "retry_warning",
			text:  "[worker work#3] 재시도 대기 중",
			level: "warn",
			tag:   "worker work#3",
		},
		{
			name:  "error_precedes_retry_warning",
			text:  "[task-llm] 작업 2 비스트리밍 호출 실패, 1s 후 동일 공급자로 재시도 (1/3)",
			level: "error",
			tag:   "task-llm",
		},
		{
			name:  "english_error",
			text:  "[worker work#3] model ERROR: connection reset",
			level: "error",
			tag:   "worker work#3",
		},
		{
			name:  "legacy_failure",
			text:  "[notify] \u5931\u8d25: connection reset",
			level: "error",
			tag:   "notify",
		},
		{
			name:  "legacy_retry",
			text:  "[notify] \u91cd\u8bd5 (1/3)",
			level: "warn",
			tag:   "notify",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLog("2026/10/27 01:09:43 " + tc.text)
			if got.Level != tc.level {
				t.Errorf("level=%q, want %q", got.Level, tc.level)
			}
			if got.Tag != tc.tag {
				t.Errorf("tag=%q, want %q", got.Tag, tc.tag)
			}
			if got.Text != tc.text {
				t.Errorf("text=%q, want %q", got.Text, tc.text)
			}
		})
	}
}
