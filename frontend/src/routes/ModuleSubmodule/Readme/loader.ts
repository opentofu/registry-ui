import { queryClient } from "@/query";
import { getModuleSubmoduleReadmeQuery } from "../query";
import { defer, LoaderFunction } from "react-router-dom";
import { ModuleRouteContext } from "@/routes/Module/types";

export const moduleSubmoduleReadmeLoader: LoaderFunction = async (
  { params },
  context,
) => {
  const readme = queryClient.ensureQueryData(
    getModuleSubmoduleReadmeQuery(
      params.namespace,
      params.name,
      params.target,
      (context as ModuleRouteContext).version,
      params.submodule,
    ),
  );

  return defer({
    readme,
  });
};
