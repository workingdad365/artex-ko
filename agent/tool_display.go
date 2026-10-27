package agent

import (
	"strings"

	"github.com/Autumn-27/artex/guard"
)

// SDK 오류의 표시 문구만 변환하며 모델 입력, 감사 기록, 셸 출력은 원문을 유지한다.
func toolResultDisplay(toolName, output string, isError bool) string {
	if !isError {
		return output
	}
	if shown := guard.DisplayMessage(output); shown != output {
		return shown
	}
	switch toolName {
	case "shell_open", "shell_send", "shell_read", "shell_close", "shell_list":
	default:
		return output
	}
	switch output {
	case "交互式 shell 已被全局停用(AGENT_CORE_DISABLE_INTERACTIVE_SHELL)":
		return "대화형 shell이 전역으로 비활성화되었습니다(AGENT_CORE_DISABLE_INTERACTIVE_SHELL)"
	case "交互式 shell 需要后台任务管理器(未启用)":
		return "대화형 shell에 필요한 백그라운드 작업 관리자가 활성화되지 않았습니다"
	case "command 必填":
		return "command는 필수입니다"
	case "mode 需为 interactive | hands-free | dispatch | monitor":
		return "mode는 interactive | hands-free | dispatch | monitor 중 하나여야 합니다"
	case "monitor 模式必须给 watch(要盯的关键词/正则)":
		return "monitor 모드에는 watch(감시할 키워드/정규식)가 필요합니다"
	case "没有可发送的输入(text/submit/keys/hex/paste 至少给一个)":
		return "전송할 입력이 없습니다(text/submit/keys/hex/paste 중 하나 이상 필요)"
	case "开启会话失败: 交互式 shell(PTY) 在当前平台不支持":
		return "세션 열기 실패: 현재 플랫폼은 대화형 shell(PTY)을 지원하지 않습니다"
	}
	if exit, ok := strings.CutPrefix(output, "会话已结束(exit "); ok {
		if exit, ok = strings.CutSuffix(exit, ")，用 shell_read 读取尾部输出或 shell_close"); ok {
			return "세션이 종료되었습니다(exit " + exit + "). shell_read로 마지막 출력을 읽거나 shell_close를 사용하세요"
		}
	}
	for _, pair := range [][2]string{
		{"watch 正则无效 ", "watch 정규식이 올바르지 않습니다 "},
		{"prompt_regex 无效: ", "prompt_regex가 올바르지 않습니다: "},
		{"开启会话失败: ", "세션 열기 실패: "},
		{"会话不存在: ", "세션이 없습니다: "},
		{"未知按键: ", "알 수 없는 키: "},
		{"写入失败: ", "쓰기 실패: "},
	} {
		if rest, ok := strings.CutPrefix(output, pair[0]); ok {
			return pair[1] + rest
		}
	}
	return output
}
