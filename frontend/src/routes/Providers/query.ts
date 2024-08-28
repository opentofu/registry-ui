import { queryOptions } from "@tanstack/react-query";
import { definitions } from "@/api";
import { api } from "@/query";

export const getProvidersQuery = () =>
  queryOptions({
    queryKey: ["providers"],
    queryFn: async () => {
      const data =
        await api(`providers/index.json`).json<definitions["ProviderList"]>();

      return data.providers;
    },
  });
