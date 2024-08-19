import { definitions } from "@/api";
import { queryOptions, skipToken } from "@tanstack/react-query";

export const getProviderVersionDataQuery = (
  namespace: string | undefined,
  provider: string | undefined,
  version: string | undefined,
) => {
  const hasParams = namespace && provider && version;

  return queryOptions({
    queryKey: ["provider-version", namespace, provider, version],
    queryFn: hasParams
      ? async () => {
          const response = await fetch(
            `${import.meta.env.VITE_DATA_API_URL}/providers/${namespace}/${provider}/${version}/index.json`,
          );

          const data = await response.json();

          return data as definitions["ProviderVersion"];
        }
      : skipToken,
  });
};

export const getProviderDocsQuery = (
  namespace: string | undefined,
  provider: string | undefined,
  version: string | undefined,
  type: string | undefined,
  name: string | undefined,
  lang: string | null,
) => {
  return queryOptions({
    queryKey: ["provider-doc", namespace, provider, type, name, lang, version],
    queryFn: async () => {
      try {
        const urlBase = `${import.meta.env.VITE_DATA_API_URL}/providers/${namespace}/${provider}/${version}`;
        const requestURL =
          type === undefined && name === undefined
            ? `${urlBase}/index.md`
            : `${urlBase}/${lang ? `cdktf/${lang}/` : ""}${type}/${name}.md`;

        const response = await fetch(requestURL);
        return response.text();
      } catch {
        return "";
      }
    },
  });
};

export const getProviderDataQuery = (
  namespace: string | undefined,
  provider: string | undefined,
) => {
  return queryOptions({
    queryKey: ["provider", namespace, provider],
    queryFn:
      namespace && provider
        ? async () => {
            const response = await fetch(
              `${import.meta.env.VITE_DATA_API_URL}/providers/${namespace}/${provider}/index.json`,
            );

            const data = await response.json();

            return data as definitions["Provider"];
          }
        : skipToken,
  });
};
