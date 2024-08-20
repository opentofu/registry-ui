import { queryClient } from "@/query";
import { getModuleVersionDataQuery } from "./query";
import { defer, LoaderFunction } from "react-router-dom";
import { ModuleRouteContext } from "./types";

export const moduleLoader: LoaderFunction = async ({ params }, context) => {
  const versionData = queryClient.ensureQueryData(
    getModuleVersionDataQuery(
      params.namespace,
      params.name,
      params.target,
      (context as ModuleRouteContext).version,
    ),
  );

  return defer({
    versionData,
  });
};
