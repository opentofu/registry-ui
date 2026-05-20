import { components } from "@/api";

type DocType = keyof Omit<components["schemas"]["ProviderDocs"], "index">;

const docsTypes: Array<DocType> = [
  "resources",
  "datasources",
  "functions",
  "guides",
];

export function isValidDocsType(type: string | undefined): type is DocType {
  return !!type && (docsTypes as string[]).includes(type);
}
