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
