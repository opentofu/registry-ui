import { queryOptions } from "@tanstack/react-query";
import lunr from "lunr";

export const getSearchIndexQuery = () =>
  queryOptions({
    queryKey: ["search-index"],
    queryFn: async () => {
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/search.json`,
      );

      const res = await response.json();

      return lunr.Index.load(res);
    },
  });
