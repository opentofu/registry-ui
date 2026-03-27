import { definitions } from "@/api";

export function getDocumentationUrl(result: definitions["SearchResultItem"]): string {
  const vars = result.link_variables;
  
  switch (result.type) {
    case "module":
      return `/module/${vars.namespace}/${vars.name}/${vars.target_system}/${vars.version || "latest"}`;
    case "provider":
      return `/provider/${vars.namespace}/${vars.name}/${vars.version || "latest"}`;
    case "provider/resource":
      return `/provider/${vars.namespace}/${vars.name}/${vars.version || "latest"}/docs/resources/${vars.id}`;
    case "provider/datasource":
      return `/provider/${vars.namespace}/${vars.name}/${vars.version || "latest"}/docs/datasources/${vars.id}`;
    case "provider/function":
      return `/provider/${vars.namespace}/${vars.name}/${vars.version || "latest"}/docs/functions/${vars.id}`;
    default:
      return "";
  }
}