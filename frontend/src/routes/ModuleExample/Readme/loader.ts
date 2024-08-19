import { queryClient } from "@/query";
import { getModuleExampleReadmeQuery } from "../query";
import { defer, LoaderFunction } from "react-router-dom";
import { ModuleRouteContext } from "@/routes/Module/types";

export const moduleExampleReadmeLoader: LoaderFunction = async (
  { params },
  context,
) => {
  const readme = queryClient.ensureQueryData(
    getModuleExampleReadmeQuery(
      params.namespace,
      params.name,
      params.target,
      (context as ModuleRouteContext).version,
      params.example,
    ),
  );

  return defer({
    readme,
  });
};
