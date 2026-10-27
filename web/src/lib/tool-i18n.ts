import catalog from "@/lib/tool-descriptions.ko.json";

// 긴 설명부터 치환하여 공통 필드명이 설명의 일부를 먼저 바꾸지 않도록 한다.
const translations = Object.entries(catalog).sort(([a], [b]) => b.length - a.length);

export function toolText(text: string, locale: string, system = true): string {
  if (locale !== "ko" || !system) return text;
  let result = text;
  for (const [source, translated] of translations) result = result.split(source).join(translated);
  return result;
}

// 편집하지 않은 표시 번역은 저장 시 원문을 유지한다.
export function toolTextForSave(edited: string, original: string, locale: string): string {
  return edited === toolText(original, locale) ? original : edited;
}
