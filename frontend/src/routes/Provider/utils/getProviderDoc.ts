import { definitions } from "@/api";
import { isValidDocsType } from "./isValidDocsType";
import { isValidCDKTFLang } from "./isValidCDKTFLang";

export function getProviderDoc(
  providerVersionData: definitions["ProviderVersion"],
  type: string | undefined,
  doc: string | undefined,
  lang: string | null,
) {
  const docs = isValidCDKTFLang(lang)
    ? providerVersionData.cdktf_docs[lang]
    : providerVersionData.docs;

  if (!type && !doc) {
    return docs.index;
  }

  if (doc && isValidDocsType(type)) {
    return docs?.[type]?.find((d) => d.name === doc);
  }
}
