package server

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/traffic"
)

// 실제 조립된 도구 설명과 중첩 스키마까지 한국어 표시 카탈로그의 누락을 검사한다.
func TestBuiltinToolKoreanDisplayCoverage(t *testing.T) {
	raw, err := os.ReadFile("../web/src/lib/tool-descriptions.ko.json")
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]string
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(catalog))
	for key, value := range catalog {
		keys = append(keys, key)
		if strings.TrimSpace(value) == "" || strings.ContainsFunc(value, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Errorf("invalid Korean translation: %q", value)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	check := func(path, text string) {
		for _, key := range keys {
			text = strings.ReplaceAll(text, key, catalog[key])
		}
		if strings.ContainsFunc(text, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Errorf("untranslated %s: %s", path, text)
		}
	}
	var schema func(string, any)
	schema = func(path string, value any) {
		switch v := value.(type) {
		case map[string]any:
			for key, child := range v {
				if key == "description" {
					if text, ok := child.(string); ok {
						check(path+"."+key, text)
					}
				} else {
					schema(path+"."+key, child)
				}
			}
		case []any:
			for _, child := range v {
				schema(path, child)
			}
		}
	}
	for _, seed := range agent.BuiltinToolSeeds() {
		check(seed.Key, seed.Desc)
		schema(seed.Key, seed.Schema)
	}
	s := &Server{}
	tools := append(s.orchestrationTools(), s.platformTools()...)
	tools = append(tools, s.findingRetestTools()...)
	tools = append(tools, traffic.SeedToolMetas()...)
	for _, tool := range tools {
		check(tool.Name(), tool.Description())
		schema(tool.Name(), tool.InputSchema())
	}
}
