import { queryOptions } from "@tanstack/react-query";
import { definitions } from "@/api";

export const getProvidersQuery = () =>
  queryOptions({
    queryKey: ["providers"],
    queryFn: async () => {
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/providers/index.json`,
      );

      const res = await response.json();

      return res.providers as definitions["ProviderList"]["providers"];
    },
  });

export const getNamespaceProvidersQuery = (namespace: string | undefined) =>
  queryOptions({
    queryKey: ["namespace-providers", namespace],
    queryFn: async () => {
      // TODO: Handle fetching providers by individual namespace
      // For now this is commented out until we have a better understanding of how to handle this
      // const response = await fetch(
      //   `${import.meta.env.VITE_DATA_API_URL}/mega_index.json`,
      // );

      // const res = await response.json();

      return [];
    },
  });
