import { useLoaderData, useParams, useSearchParams } from "react-router";
import { ProviderRouteContext } from "../types";

export function useProviderParams() {
  const { version, rawVersion } = useLoaderData() as ProviderRouteContext;

  const { namespace, provider, type, doc } = useParams<{
    namespace: string;
    provider: string;
    type: string;
    doc: string;
  }>();

  const [searchParams] = useSearchParams();

  return {
    version,
    namespace,
    provider,
    type,
    doc,
    lang: searchParams.get("lang"),
    isLatest: rawVersion === "latest",
  };
}
