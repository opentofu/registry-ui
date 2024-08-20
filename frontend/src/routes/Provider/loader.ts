import { queryClient } from "@/query";
import { getProviderVersionDataQuery } from "./query";
import { defer, LoaderFunction } from "react-router-dom";
import { ProviderRouteContext } from "./types";

export const providerLoader: LoaderFunction = async ({ params }, context) => {
  const versionData = queryClient.ensureQueryData(
    getProviderVersionDataQuery(
      params.namespace,
      params.provider,
      (context as ProviderRouteContext).version,
    ),
  );

  return defer({
    versionData,
  });
};
