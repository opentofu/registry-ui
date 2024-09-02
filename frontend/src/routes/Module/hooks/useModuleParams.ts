import { useParams, useRouteLoaderData } from "react-router-dom";
import { ModuleRouteContext } from "../types";

export function useModuleParams() {
  const { version, rawVersion } = useRouteLoaderData(
    "module-version",
  ) as ModuleRouteContext;

  const { namespace, name, target } = useParams<{
    namespace: string;
    name: string;
    target: string;
  }>();

  return {
    version,
    namespace,
    name,
    target,
    isLatest: rawVersion === "latest",
  };
}
