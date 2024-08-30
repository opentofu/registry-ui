import { queryOptions, skipToken } from "@tanstack/react-query";
import { api } from "./query";
import { ApiSearchResult } from "./components/Search/types";

export const getSearchQuery = (query: string) =>
  queryOptions({
    queryKey: ["search", query],
    queryFn:
      query.length > 0
        ? async ({ signal }) => {
            const response = await api(
              `/search?q=${encodeURIComponent(query)}`,

              {
                signal,
              },
            );

            const res = await response.json();
            return res as ApiSearchResult[];
          }
        : skipToken,
  });
