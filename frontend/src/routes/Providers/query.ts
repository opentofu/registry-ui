import { queryOptions } from "@tanstack/react-query";
import { components } from "@/api";
import { api } from "@/query";

export const getProvidersQuery = () =>
  queryOptions({
    queryKey: ["providers"],
    queryFn: async () => {
      const data =
        await api(`providers/index.json`).json<components["schemas"]["ProviderList"]>();

      return data.providers;
    },
  });
