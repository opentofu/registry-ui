import { queryOptions, skipToken } from "@tanstack/react-query";
import { api } from "@/query";
import { definitions } from "./api";

export const getSearchQuery = (query: string) =>
  queryOptions({
    queryKey: ["search", query],
    queryFn:
      query.length > 0
        ? async ({ signal }) => {
            // This is the lazy man's debounce, by waiting 100ms before making the request
            // and utilizing the signal to cancel the request if the query changes
            // we get a debounce effect
            await (() => new Promise((resolve) => setTimeout(resolve, 100)))();
            if (signal.aborted) {
              return;
            }

            return await api(`search?q=${encodeURIComponent(query)}`, {
              signal,
            }).json<definitions["SearchResultItem"][]>();
          }
        : skipToken,
  });
