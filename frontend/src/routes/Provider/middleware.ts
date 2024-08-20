import { queryClient } from "@/query";
import { LoaderFunction, redirect } from "react-router-dom";
import { getProviderDataQuery } from "./query";
import { ProviderRouteContext } from "./types";

export const providerMiddleware: LoaderFunction = async (
  { params },
  context,
) => {
  const { namespace, provider, version } = params;

  const data = await queryClient.ensureQueryData(
    getProviderDataQuery(namespace, provider),
  );

  const [latestVersion] = data.versions;

  if (version === latestVersion.id || !version) {
    return redirect(`/provider/${namespace}/${provider}/latest`);
  }

  const providerContext = context as ProviderRouteContext;

  providerContext.version = version === "latest" ? latestVersion.id : version;
  providerContext.rawVersion = version;
  providerContext.namespace = namespace;
  providerContext.provider = provider;
};
