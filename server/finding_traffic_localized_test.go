package server

import "testing"

// finding_traffic.go 의 사용자 노출 HTTP 에러 응답 문구를 한국어로 유지하는 회귀 방어
// 테스트다. 한국어 판정은 F3a 가 만든 assertKoreanError(한글 포함·중국어 한자 0)를
// 재사용한다. 네 문구를 반환하는 경로(findingTrafficAccess·bindFindingTraffic·
// editFindingTraffic)는 전부 s.m.pg.GetFinding(DB) 게이트 뒤라 DB 없는 이 호스트에서
// 끝까지 돌 수 없으므로 상수 자체를 핀 고정한다(goals_api·notify_api 의 DB 게이트
// 경로와 같은 방식). 누군가 이 리터럴을 중국어로 되돌리면 이 테스트가 실패한다.
//
// readEvidencePreview의 오류와 이진 본문 안내도 한국어로 제공한다.
// 모델용 도구 정의는 유지하고 웹 카탈로그에서 설명을 번역한다.
func TestFindingTrafficErrorsLocalized(t *testing.T) {
	for _, c := range []struct {
		label, msg string
	}{
		{"inherited_readonly", errFindingTrafficInheritedReadonly},
		{"select_required", errFindingTrafficSelectRequired},
		{"version_required", errFindingTrafficVersionRequired},
		{"binding_ids_required", errFindingTrafficBindingIDsRequired},
	} {
		assertKoreanError(t, c.label, c.msg)
	}
}
