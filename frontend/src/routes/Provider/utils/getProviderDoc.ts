import { definitions } from "@/api";
import { isValidDocsType } from "./isValidDocsType";

export function getProviderDoc(
  docs: definitions["ProviderDocs"],
  type: string | undefined,
  doc: string | undefined,
) {
  if (!type && !doc) {
    return docs.index;
  }

  if (doc && isValidDocsType(type)) {
    return docs?.[type]?.find((d) => d.name === doc);
  }
}
