import { queryClient } from "@/query";
import { getModuleReadmeQuery } from "../query";
import { defer, LoaderFunction } from "react-router-dom";
import { ModuleRouteContext } from "../types";

export const moduleReadmeLoader: LoaderFunction = async (
  { params },
  context,
) => {
  const readme = queryClient.ensureQueryData(
    getModuleReadmeQuery(
      params.namespace,
      params.name,
      params.target,
      (context as ModuleRouteContext).version,
    ),
  );

  return defer({
    readme,
  });
};
