import { queryClient } from "@/query";
import { defer, LoaderFunction } from "react-router-dom";
import { getModuleExampleDataQuery } from "./query";
import { ProviderRouteContext } from "../Provider/types";

export const moduleExampleLoader: LoaderFunction = async (
  { params },
  context,
) => {
  const example = queryClient.ensureQueryData(
    getModuleExampleDataQuery(
      params.namespace,
      params.name,
      params.target,
      (context as ProviderRouteContext).version,
      params.example,
    ),
  );

  return defer({
    example,
  });
};
