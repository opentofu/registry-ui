import { queryOptions } from "@tanstack/react-query";
import lunr from "lunr";
import { api } from "./query";

export const getSearchIndexQuery = () =>
  queryOptions({
    queryKey: ["search-index"],
    queryFn: async () => {
      const data = await api(`search.json`).json();

      return lunr.Index.load(data);
    },
  });
