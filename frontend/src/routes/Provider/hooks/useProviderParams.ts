import { useLoaderData, useParams, useSearchParams } from "react-router-dom";

export function useProviderParams() {
  const { version } = useLoaderData() as {
    version: string;
  };

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
  };
}
