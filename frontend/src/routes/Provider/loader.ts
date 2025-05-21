import { queryClient } from "@/query";
import { getProviderVersionDataQuery } from "./query";
import { LoaderFunction } from "react-router";
import { ProviderRouteContext } from "./types";

export const providerLoader: LoaderFunction = async ({ params }, context) => {
  const versionData = queryClient.ensureQueryData(
    getProviderVersionDataQuery(
      params.namespace,
      params.provider,
      (context as ProviderRouteContext).version,
    ),
  );

  return {
    versionData,
  };
};
