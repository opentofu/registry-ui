import { LoaderFunction } from "react-router";
import { getProviderDocsQuery } from "../query";
import { queryClient } from "@/query";
import { ProviderRouteContext } from "../types";

export const providerOverviewLoader: LoaderFunction = (
  { params, request },
  context,
) => {
  const url = new URL(request.url);
  const lang = url.searchParams.get("lang");

  const docs = queryClient.ensureQueryData(
    getProviderDocsQuery(
      params.namespace,
      params.provider,
      (context as ProviderRouteContext).version,
      undefined,
      undefined,
      lang,
    ),
  );

  return {
    docs,
  };
};
