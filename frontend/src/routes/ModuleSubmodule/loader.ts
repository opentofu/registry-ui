import { queryClient } from "@/query";
import { LoaderFunction } from "react-router";
import { getModuleSubmoduleDataQuery } from "./query";
import { ProviderRouteContext } from "../Provider/types";

export const moduleSubmoduleLoader: LoaderFunction = async (
  { params },
  context,
) => {
  const submodule = queryClient.ensureQueryData(
    getModuleSubmoduleDataQuery(
      params.namespace,
      params.name,
      params.target,
      (context as ProviderRouteContext).version,
      params.submodule,
    ),
  );

  return {
    submodule,
  };
};
