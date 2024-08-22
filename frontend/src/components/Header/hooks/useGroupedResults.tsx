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

const typeSortingOrder = {
  [SearchRefType.Provider]: 0,
  [SearchRefType.Module]: 1,
  [SearchRefType.ProviderResource]: 2,
  [SearchRefType.ProviderDatasource]: 3,
  [SearchRefType.ProviderFunction]: 4,
  [SearchRefType.Other]: 5,
};

export const getTypeLabel = (type: SearchRefType) => {
  switch (type) {
    case SearchRefType.Module:
      return "Module";
    case SearchRefType.Provider:
    case SearchRefType.ProviderResource:
    case SearchRefType.ProviderDatasource:
    case SearchRefType.ProviderFunction:
      return "Provider";
    case SearchRefType.Other:
      return "Other";
  }
};

const parseRef = (ref: string): SearchRef => {
  return JSON.parse(ref);
};

const useGroupedResults = (
  deferredQuery: string,
  data: Index | undefined,
): Array<
  [SearchRefType, Array<SearchRef & { ref: string; typeLabel: string }>]
> => {
  return useMemo(() => {
    if (deferredQuery === "" || !data) {
      return [];
    }

    const results = data.search(deferredQuery).slice(0, 10);

    const groupedResults: Map<
      SearchRefType,
      Array<SearchRef & { ref: string; typeLabel: string }>
    > = new Map();

    results.forEach((r) => {
      const parsed = parseRef(r.ref);

      if (!groupedResults.has(parsed.type)) {
        groupedResults.set(parsed.type, []);
      }

      groupedResults.get(parsed.type)?.push({
        ...parsed,
        ref: r.ref,
        typeLabel: getTypeLabel(parsed.type),
      });
    });

    return Array.from(groupedResults.entries()).sort(
      ([a], [b]) => typeSortingOrder[a] - typeSortingOrder[b],
    );
  }, [deferredQuery, data]);
};

export default useGroupedResults;
