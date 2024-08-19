import { queryClient } from "@/query";
import { LoaderFunction, redirect } from "react-router-dom";
import { getProviderDataQuery } from "./query";
import { ProviderRouteContext } from "./types";

export const providerMiddleware: LoaderFunction = async (
  { params },
  context,
) => {
  const data = await queryClient.ensureQueryData(
    getProviderDataQuery(params.namespace, params.provider),
  );

  const [latestVersion] = data.versions;

  if (params.version === latestVersion.id || !params.version) {
    return redirect(`/provider/${params.namespace}/${params.provider}/latest`);
  }

  (context as ProviderRouteContext).version =
    params.version === "latest" ? latestVersion.id : params.version;
};
