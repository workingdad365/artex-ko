import catalog from "@/lib/skill-descriptions.ko.json";

// 배포된 기본 설명과 정확히 일치할 때만 표시용 번역을 적용한다.
export function skillDescription(description: string, locale: string): string {
  if (locale !== "ko") return description;
  return (catalog as Record<string, string>)[description] ?? description;
}
