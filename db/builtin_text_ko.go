package db

import "strings"

// 기존 기본값과 정확히 일치하는 표시 문구만 갱신한다.
var builtinTextKO = map[string]string{
	"目标拆解": "목표 분해",
	"把渗透任务目标拆解成若干独立、可验证的子目标。": "침투 테스트 목표를 독립적으로 검증할 수 있는 하위 목표로 분해합니다.",
	"任务描述（测试对象/背景）":           "작업 설명(테스트 대상/배경)",
	"测试 example.com 站点":       "example.com 사이트 테스트",
	"规划":                      "계획",
	"读取态势、判定目标，只在确有未覆盖的新方向时补充探索意图（每任务一个规划循环）。": "현황과 목표 달성 여부를 확인하고 미탐색 방향이 있을 때 탐색 의도를 추가합니다(작업별 계획 루프).",
	"任务总目标":                 "전체 작업 목표",
	"拿下 example.com 的管理员权限": "example.com의 관리자 권한 검증",
	"资产计数/类型分布摘要(可选)":       "자산 수/유형별 분포 요약(선택)",
	"主": "메인",
	"人机接口：观察进展，把人的意图落成 hint 或高优先级意图。": "진행 상황을 확인하고 사용자 요청을 힌트 또는 우선순위가 높은 의도로 반영합니다.",
	"当前任务目标":      "현재 작업 목표",
	"开局态势摘要(可选)":  "초기 현황 요약(선택)",
	"已确认漏洞摘要(可选)": "확인된 취약점 요약(선택)",
	"执行":          "실행",
	"领取一条意图执行，把发现的事实/漏洞写回知识图谱后停止。": "의도 하나를 실행하고 발견한 사실과 취약점을 지식 그래프에 기록한 뒤 종료합니다.",
	"记录代理地址(驱动 if 双文案)":            "기록 프록시 주소(조건부 안내에 사용)",
	"worker 自我标识(可选)":              "Worker 식별자(선택)",
	"平台操作助手：用工具管理任务(建/看/暂停/给提示)与资产，并可创建/修改 skill、自定义工具、MCP。": "도구로 작업과 자산을 관리하고 스킬, 사용자 정의 도구, MCP를 생성하거나 수정하는 플랫폼 도우미입니다.",
	"渗透测试": "침투 테스트",
	"独立渗透 agent：一人从侦察→找攻击面→深入利用→验证→收尾走完整条链，自己规划、自己执行、自己对抗式验证。": "정찰, 공격 표면 탐색, 심층 테스트, 검증, 마무리까지 직접 계획하고 실행하는 독립 침투 테스트 Agent입니다.",
	"[内置] 政府网站 (.gov)":    "[내장] 정부 사이트 (.gov)",
	"[内置] 政府网站 (.gov.cn)": "[내장] 정부 사이트 (.gov.cn)",
	"[内置] 教育网站 (.edu)":    "[내장] 교육 사이트 (.edu)",
	"[内置] 教育网站 (.edu.cn)": "[내장] 교육 사이트 (.edu.cn)",
	"[内置] 递归强制删除 rm -rf":  "[내장] 재귀 강제 삭제 rm -rf",
	"禁止执行递归强制删除（rm -rf / rm --recursive），可能永久损坏系统或靶机环境": "시스템 또는 테스트 대상 환경을 영구 손상할 수 있는 재귀 강제 삭제(rm -rf / rm --recursive)를 금지합니다",
	"[内置] 删除系统关键目录":                                              "[내장] 시스템 핵심 디렉터리 삭제",
	"禁止删除系统关键路径":                                                 "시스템 핵심 경로 삭제를 금지합니다",
	"[内置] 磁盘格式化 mkfs":                                            "[내장] 디스크 포맷 mkfs",
	"禁止格式化磁盘（mkfs）":                                              "디스크 포맷(mkfs)을 금지합니다",
	"[内置] 覆写磁盘设备 dd":                                             "[내장] 디스크 장치 덮어쓰기 dd",
	"禁止使用 dd 覆写磁盘设备":                                             "dd를 사용한 디스크 장치 덮어쓰기를 금지합니다",
	"[内置] Fork 炸弹":                                               "[내장] Fork 폭탄",
	"禁止执行 Fork 炸弹":                                               "Fork 폭탄 실행을 금지합니다",
	"[内置] 关机 / 重启":                                               "[내장] 종료 / 재부팅",
	"禁止执行关机或重启命令":                                                "종료 또는 재부팅 명령 실행을 금지합니다",
	"[内置] 杀死全部进程":                                                "[내장] 전체 프로세스 종료",
	"禁止 kill -9 -1 或 killall -9（杀死所有进程）":                         "kill -9 -1 또는 killall -9(전체 프로세스 종료)를 금지합니다",
	"[内置] 磁盘擦除 shred / wipe":                                     "[내장] 디스크 지우기 shred / wipe",
	"禁止对磁盘设备执行 shred/wipe 擦除":                                    "디스크 장치에 대한 shred/wipe 지우기를 금지합니다",
	"[内置] 清空防火墙规则":                                               "[내장] 방화벽 규칙 초기화",
	"禁止清空防火墙规则（iptables -F / nft flush）":                         "방화벽 규칙 초기화(iptables -F / nft flush)를 금지합니다",
	"[内置] SQL DROP DATABASE / TABLE / SCHEMA":                    "[내장] SQL DROP DATABASE / TABLE / SCHEMA",
	"禁止执行 DROP 操作，可能不可逆地销毁数据库对象":                                 "DB 객체를 복구할 수 없게 삭제할 수 있는 DROP 실행을 금지합니다",
	"[内置] SQL TRUNCATE":                                          "[내장] SQL TRUNCATE",
	"禁止执行 TRUNCATE，可能清空数据表所有数据":                                  "테이블의 모든 데이터를 삭제할 수 있는 TRUNCATE 실행을 금지합니다",
	"[内置] MongoDB drop / dropDatabase":                           "[내장] MongoDB drop / dropDatabase",
	"禁止执行 MongoDB drop 操作":                                       "MongoDB drop 실행을 금지합니다",
	"[内置] Redis FLUSHALL / FLUSHDB":                              "[내장] Redis FLUSHALL / FLUSHDB",
	"禁止执行 Redis FLUSHALL / FLUSHDB，可能清空全部缓存数据":                   "전체 캐시 데이터를 삭제할 수 있는 Redis FLUSHALL / FLUSHDB 실행을 금지합니다",
	"[内置] curl / wget 发送 DELETE 请求":                              "[내장] curl / wget DELETE 요청",
	"禁止通过 curl/wget 发送 HTTP DELETE 请求，可能删除目标系统数据":                "대상 시스템의 데이터를 삭제할 수 있는 curl/wget HTTP DELETE 요청을 금지합니다",
	"[内置] Python HTTP 客户端 DELETE（requests/httpx/aiohttp）":        "[내장] Python HTTP 클라이언트 DELETE(requests/httpx/aiohttp)",
	"禁止使用 Python HTTP 客户端发送 DELETE 请求":                           "Python HTTP 클라이언트의 DELETE 요청을 금지합니다",
	"[内置] 脚本中声明 HTTP DELETE 方法（JS/通用）":                           "[내장] 스크립트의 HTTP DELETE 메서드(JS/일반)",
	"禁止在脚本中声明并发送 HTTP DELETE 请求":                                 "스크립트에서 HTTP DELETE 요청을 선언하고 전송하는 것을 금지합니다",
	"[内置] 批量清空 / 清除接口路径":                                         "[내장] 일괄 초기화 / 삭제 API 경로",
	"禁止调用批量清空或销毁类接口（/clear /wipe /flush /purge 等）":               "일괄 초기화 또는 삭제 API(/clear /wipe /flush /purge 등) 호출을 금지합니다",
	"[内置] 破坏性系统命令":                                               "[내장] 파괴적 시스템 명령",
	"破坏性命令被拒绝（rm -rf / / mkfs / dd / fork bomb / 关机重启 / 覆写磁盘设备）": "파괴적 명령이 거부되었습니다(rm -rf /, mkfs, dd, fork bomb, 종료/재부팅, 디스크 장치 덮어쓰기)",
	"[内置] 数据外泄管道":                                                "[내장] 데이터 유출 파이프",
	"疑似数据外泄管道被拒绝（命令输出经 curl/wget/nc 外传）":                         "데이터 유출이 의심되는 파이프가 거부되었습니다(curl/wget/nc로 명령 출력 외부 전송)",
	"[内置] 删除类接口路径":                                               "[내장] 삭제 API 경로",
	"禁止调用删除类接口（/delete /remove /unlink /erase 等），不论使用哪种 HTTP 方法——多数应用的删除接口用 GET/POST 就能触发，同样会真实删除目标数据": "HTTP 메서드와 관계없이 삭제 API(/delete /remove /unlink /erase 등) 호출을 금지합니다. GET/POST로도 삭제가 실행되어 실제 대상 데이터가 손실될 수 있습니다",
}

func builtinText(text string) string {
	if ko, ok := builtinTextKO[text]; ok {
		return ko
	}
	return text
}

// 이름, 설명, 메모만 갱신하며 사용자가 바꾼 값과 규칙의 실행 설정은 유지한다.
func (d *DB) localizeBuiltinText() error {
	const flag = "builtin_text_ko_v1"
	if v, _, err := d.GetSetting(flag); err != nil {
		return err
	} else if v == "done" {
		return nil
	}
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// 규칙 식별은 갱신 전 이름으로 고정하여 문구 변경 순서에 의존하지 않는다.
	if _, err = tx.Exec(`CREATE TEMP TABLE ko_builtin_rule_ids ON COMMIT DROP AS
		SELECT id FROM intercept_rules WHERE name LIKE '[内置] %' OR name LIKE '[내장] %'`); err != nil {
		return err
	}
	for old, ko := range builtinTextKO {
		for _, query := range []string{
			`UPDATE agents SET name=$2 WHERE builtin AND name=$1`,
			`UPDATE agents SET description=$2 WHERE builtin AND description=$1`,
			`UPDATE agent_prompt_vars SET description=$2 WHERE description=$1 AND agent_id IN (SELECT id FROM agents WHERE builtin)`,
			`UPDATE agent_prompt_vars SET example=$2 WHERE example=$1 AND agent_id IN (SELECT id FROM agents WHERE builtin)`,
			`UPDATE asset_intercept_rules SET note=$2 WHERE builtin AND note=$1`,
			`UPDATE intercept_rules SET message=$2 WHERE message=$1 AND id IN (SELECT id FROM ko_builtin_rule_ids)`,
		} {
			if _, err = tx.Exec(query, old, ko); err != nil {
				return err
			}
		}
		if strings.HasPrefix(old, "[内置] ") {
			if _, err = tx.Exec(`UPDATE intercept_rules SET name=$2 WHERE name=$1 AND id IN (SELECT id FROM ko_builtin_rule_ids)`, old, ko); err != nil {
				return err
			}
		}
	}
	if _, err = tx.Exec(`UPDATE agent_prompts SET note=CASE note WHEN '内置默认' THEN '내장 기본값' ELSE '내장 기본값으로 복원' END
		WHERE updated_by='system' AND note IN ('内置默认','恢复为内置默认')`); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES ($1,'done') ON CONFLICT(key) DO UPDATE SET value='done'`, flag); err != nil {
		return err
	}
	return tx.Commit()
}
