import { useParams, useRouteLoaderData } from "react-router-dom";

export function useModuleParams() {
  const { version } = useRouteLoaderData("module-version") as {
    version: string;
  };

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
  };
}
