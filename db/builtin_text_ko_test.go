package db

import (
	"strings"
	"testing"
	"unicode"
)

func TestBuiltinTextMigrationPreservesUserState(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	for _, builtin := range builtinAgents {
		var name, description string
		if err := d.QueryRow(`SELECT name,description FROM agents WHERE key=$1`, builtin.key).Scan(&name, &description); err != nil {
			t.Fatal(err)
		}
		if strings.ContainsFunc(name+description, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Fatalf("untranslated builtin %s: %q %q", builtin.key, name, description)
		}
	}
	var id int64
	if err := d.QueryRow(`INSERT INTO intercept_rules(name,enabled,priority,match_target,match_type,pattern,action,message)
		VALUES ('[内置] Fork 炸弹',false,123,'tool_input','string','fixture-only','ask','사용자 수정 메시지') RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = d.Exec(`DELETE FROM intercept_rules WHERE id=$1`, id) })
	var originalName, originalDescription string
	if err := d.QueryRow(`SELECT name,description FROM agents WHERE key='planner'`).Scan(&originalName, &originalDescription); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = d.Exec(`UPDATE agents SET name=$1,description=$2 WHERE key='planner'`, originalName, originalDescription)
	})
	if _, err := d.Exec(`UPDATE agents SET name='사용자 플래너',description='사용자 설명' WHERE key='planner'`); err != nil {
		t.Fatal(err)
	}
	var varID int64
	var originalVarDescription, originalExample string
	if err := d.QueryRow(`SELECT v.id,v.description,v.example FROM agent_prompt_vars v JOIN agents a ON a.id=v.agent_id
		WHERE a.key='goals' AND v.var_name='EngagementDescription'`).Scan(&varID, &originalVarDescription, &originalExample); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = d.Exec(`UPDATE agent_prompt_vars SET description=$2,example=$3 WHERE id=$1`, varID, originalVarDescription, originalExample)
	})
	if _, err := d.Exec(`UPDATE agent_prompt_vars SET description='사용자 변수 설명',example='测试 example.com 站点' WHERE id=$1`, varID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`DELETE FROM settings WHERE key='builtin_text_ko_v1'`); err != nil {
		t.Fatal(err)
	}
	if err := d.localizeBuiltinText(); err != nil {
		t.Fatal(err)
	}
	var name, message, action, pattern string
	var enabled bool
	var priority int
	if err := d.QueryRow(`SELECT name,message,action,pattern,enabled,priority FROM intercept_rules WHERE id=$1`, id).
		Scan(&name, &message, &action, &pattern, &enabled, &priority); err != nil {
		t.Fatal(err)
	}
	if name != "[내장] Fork 폭탄" || message != "사용자 수정 메시지" || action != "ask" || pattern != "fixture-only" || enabled || priority != 123 {
		t.Fatalf("migration changed user state: %q %q %q %q %v %d", name, message, action, pattern, enabled, priority)
	}
	if _, err := d.Exec(`DELETE FROM intercept_rules WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	var before, after int
	if err := d.QueryRow(`SELECT count(*) FROM intercept_rules`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := d.seedBuiltins(); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT count(*) FROM intercept_rules`).Scan(&after); err != nil || before != after {
		t.Fatalf("restart recreated rules: before=%d after=%d err=%v", before, after, err)
	}
	if err := d.QueryRow(`SELECT name,description FROM agents WHERE key='planner'`).Scan(&name, &message); err != nil {
		t.Fatal(err)
	}
	if name != "사용자 플래너" || message != "사용자 설명" {
		t.Fatalf("restart overwrote agent edits: %q %q", name, message)
	}
	if err := d.QueryRow(`SELECT description,example FROM agent_prompt_vars WHERE id=$1`, varID).Scan(&name, &message); err != nil {
		t.Fatal(err)
	}
	if name != "사용자 변수 설명" || message != "example.com 사이트 테스트" {
		t.Fatalf("default example or customized variable changed incorrectly: %q %q", name, message)
	}
}
