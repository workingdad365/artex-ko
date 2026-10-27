import { describe, expect, it } from "vitest";

import { toolText, toolTextForSave } from "./tool-i18n";

describe("도구 설명 표시와 저장", () => {
  it("기본 설명의 연결된 문장을 한국어로 표시하고 무편집 저장 시 원문을 유지한다", () => {
    const original =
      "批量登记新发现的资产，一次可混合多种类型（type 见枚举）。\nauth/technologies/params 为追加合并(append)，不覆盖原值。";
    const translated = toolText(original, "ko");
    expect(translated).not.toMatch(/\p{Script=Han}/u);
    expect(translated).toContain("auth/technologies/params");
    expect(toolTextForSave(translated, original, "ko")).toBe(original);
    expect(toolTextForSave(`${translated} 사용자 보완`, original, "ko")).toBe(`${translated} 사용자 보완`);
  });

  it("사용자 정의 도구와 다른 언어 및 알 수 없는 문구는 보존한다", () => {
    expect(toolText("任务 id", "ko", false)).toBe("任务 id");
    expect(toolText("任务 id", "zh")).toBe("任务 id");
    expect(toolText("사용자 설명 /path?arg=value", "ko")).toBe("사용자 설명 /path?arg=value");
  });
});
