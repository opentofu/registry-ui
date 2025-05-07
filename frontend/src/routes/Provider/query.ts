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

// Custom error class for documentation not found errors
export class DocNotFoundError extends Error {
  statusCode: number;

  constructor(message: string, statusCode: number) {
    super(message);
    this.name = "DocNotFoundError";
    this.statusCode = statusCode;
  }
}

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
      const urlBase = `providers/${namespace}/${provider}/${version}/${lang ? `cdktf/${lang}/` : ""}`;

      const path =
        type === undefined && name === undefined
          ? `index.md`
          : `${type}/${name}.md`;

      try {
        const response = await api(`${urlBase}${path}`);
        const text = await response.text();

        if (text === "") {
          throw new DocNotFoundError("Document is empty", 204);
        }

        return text;
      } catch (error) {
        // Check if this is a 404 error from ky, if it is then surface this up to the query caller
        // this way we can handle it in the UI nicely
        if (
          error instanceof Error &&
          "response" in error &&
          (error as Error & Record<"response", { status: number }>).response
            ?.status === 404
        ) {
          // Convert to our custom error type
          throw new DocNotFoundError("Document not found", 404);
        }
        // Re-throw other errors
        throw error;
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
    queryFn: async () => {
      const data = await api(
        `providers/${namespace}/${provider}/index.json`,
      ).json<definitions["Provider"]>();

      return data;
    },
  });
};
