import { definitions } from "@/api";

type DocType = keyof Omit<definitions["ProviderDocs"], "index">;

const docsTypes: Array<DocType> = [
  "resources",
  "datasources",
  "functions",
  "guides",
];

export function isValidDocsType(type: string | undefined): type is DocType {
  return !!type && (docsTypes as string[]).includes(type);
}
