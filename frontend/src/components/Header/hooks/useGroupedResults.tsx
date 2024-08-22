import { Index } from "lunr";
import { useMemo } from "react";

type SearchRefLink = {
  name: string;
  namespace: string;
  target_system?: string;
  version: string;
  id?: string;
};

enum SearchRefType {
  Module = "module",
  Provider = "provider",
  ProviderResource = "provider/resource",
  ProviderDatasource = "provider/datasource",
  ProviderFunction = "provider/function",
  Other = "other",
}

type SearchRef = {
  addr: string;
  type: SearchRefType;
  version: string;
  title: string;
  description: string;
  link: SearchRefLink;
  parent_id: string;
};

const parseRef = (ref: string): SearchRef => {
  return JSON.parse(ref);
};

const useGroupedResults = (
  deferredQuery: string,
  data: Index | undefined,
): Map<string, Array<SearchRef & { ref: string; typeLabel: string }>> => {
  return useMemo(() => {
    if (deferredQuery === "" || !data) {
      return new Map();
    }

    const results = data.search(deferredQuery).slice(0, 10);

    const groupedResults: Map<
      string,
      Array<SearchRef & { ref: string; typeLabel: string }>
    > = new Map();

    results.forEach((r) => {
      const parsed = parseRef(r.ref);
      let typeLabel = parsed.type.split("/").pop() || "";
      typeLabel = typeLabel.charAt(0).toUpperCase() + typeLabel.slice(1);

      if (!groupedResults.has(typeLabel)) {
        groupedResults.set(typeLabel, []);
      }

      groupedResults.get(typeLabel)?.push({
        ...parsed,
        ref: r.ref,
        typeLabel: typeLabel,
      });
    });

    return groupedResults;
  }, [deferredQuery, data]);
};

export default useGroupedResults;
