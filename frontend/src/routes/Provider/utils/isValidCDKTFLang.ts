const cdktfLangs = ["typescript", "python", "go", "java", "csharp"];

export function isValidCDKTFLang(lang: string | null): lang is string {
  return !!lang && cdktfLangs.includes(lang);
}
