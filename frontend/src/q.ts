import {
  keepPreviousData,
  queryOptions,
  skipToken,
} from "@tanstack/react-query";
import { api } from "@/query";
import { definitions } from "./api";

export const getSearchQuery = (query: string) =>
  queryOptions({
    queryKey: ["search", query],
    placeholderData: keepPreviousData,
    queryFn:
      query.length > 0
        ? async ({ signal }) => {
            return api(`search?q=${encodeURIComponent(query)}`, {
              signal,
            }).json<definitions["SearchResultItem"][]>();
          }
        : skipToken,
  });
