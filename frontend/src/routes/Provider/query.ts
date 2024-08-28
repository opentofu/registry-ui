import { definitions } from "@/api";
import { api } from "@/query";
import { queryOptions } from "@tanstack/react-query";

export const getProviderVersionDataQuery = (
  namespace: string | undefined,
  provider: string | undefined,
  version: string | undefined,
) => {
  return queryOptions({
    queryKey: ["provider-version", namespace, provider, version],
    queryFn: async () => {
      const data = await api(
        `providers/${namespace}/${provider}/${version}/index.json`,
      ).json<definitions["ProviderVersion"]>();

      return data;
    },
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
      const urlBase = `providers/${namespace}/${provider}/${version}`;
      const requestURL =
        type === undefined && name === undefined
          ? `${urlBase}/index.md`
          : `${urlBase}/${lang ? `cdktf/${lang}/` : ""}${type}/${name}.md`;

      const data = await api(requestURL).text();
      return data;
    },
  });
};

export const getProviderDataQuery = (
  namespace: string | undefined,
  provider: string | undefined,
) => {
  return queryOptions({
    queryKey: ["provider", namespace, provider],
    queryFn: async () => {
      const data = await api(
        `providers/${namespace}/${provider}/index.json`,
      ).json<definitions["Provider"]>();

      return data;
    },
  });
};
