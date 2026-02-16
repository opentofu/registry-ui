import { useMemo, useState } from "react";

type SortOption = "alphabetical" | "popularity" | "recent";

interface UseListingFiltersProps<T> {
  items: T[];
  searchQuery?: string;
  getNamespace: (item: T) => string;
  getName: (item: T) => string;
  getDescription: (item: T) => string;
  getPopularity: (item: T) => number;
  getLastUpdate: (item: T) => string;
  getForkOf: (item: T) => unknown;
  getDisplayName?: (item: T) => string;
}

export function useListingFilters<T>({
  items,
  searchQuery = "",
  getNamespace,
  getName,
  getDescription,
  getPopularity,
  getLastUpdate,
  getForkOf,
  getDisplayName,
}: UseListingFiltersProps<T>) {
  const [sortBy, setSortBy] = useState<SortOption>("popularity");

  const { filteredItems, stats } = useMemo(() => {
    // First filter out items without basic required data
    let filtered = items.filter((item) => {
      if (!item || !getNamespace(item) || !getName(item)) return false;
      
      // Hide opentofu namespace providers that are forks of hashicorp providers
      // This shows hashicorp/* instead of their opentofu/* forks
      // This is very hacky right now, and should be replaced once we re-work the backend to handle these cases
      if (getNamespace(item) === 'opentofu' && item.reverse_aliases && item.reverse_aliases.length > 0) {
        return false;
      }

      if (getNamespace(item) === 'terraform-providers') {
        return false;
      }
           
      if (!searchQuery) return true;
      
      const query = searchQuery.toLowerCase();
      const namespace = getNamespace(item).toLowerCase();
      const name = getName(item).toLowerCase();
      const description = (getDescription(item) ?? "").toLowerCase();
      const displayName = (getDisplayName?.(item) ?? "").toLowerCase();
      
      return (
        namespace.includes(query) ||
        name.includes(query) ||
        description.includes(query) ||
        displayName.includes(query) ||
        `${namespace}/${name}`.includes(query)
      );
    });

       // Sort filtered items
    const sorted = [...filtered].sort((a, b) => {
      switch (sortBy) {
        case "popularity":
          return getPopularity(b) - getPopularity(a);
        case "recent": {
          const aUpdate = getLastUpdate(a);
          const bUpdate = getLastUpdate(b);
          
          // Handle items without dates by putting them at the end
          if (!aUpdate && !bUpdate) return 0;
          if (!aUpdate) return 1;
          if (!bUpdate) return -1;
          
          const aDate = new Date(aUpdate).getTime();
          const bDate = new Date(bUpdate).getTime();
          
          // Handle invalid dates
          if (isNaN(aDate) && isNaN(bDate)) return 0;
          if (isNaN(aDate)) return 1;
          if (isNaN(bDate)) return -1;
          
          return bDate - aDate;
        }
        case "alphabetical":
        default: {
          const aName = `${getNamespace(a)}/${getName(a)}`.toLowerCase();
          const bName = `${getNamespace(b)}/${getName(b)}`.toLowerCase();
          return aName.localeCompare(bName);
        }
      }
    });

    const recentUpdates = [...items]
      .filter(item => {
        const lastUpdate = getLastUpdate(item);
        return lastUpdate && !isNaN(new Date(lastUpdate).getTime());
      })
      .sort((a, b) => new Date(getLastUpdate(b)).getTime() - new Date(getLastUpdate(a)).getTime())
      .slice(0, 15)
      .map(item => ({
        name: `${getNamespace(item)}/${getName(item)}`,
        version: item.versions?.[0]?.id,
        published: getLastUpdate(item),
        namespace: getNamespace(item),
        itemName: getName(item),
        target: item.addr?.target, // For modules
      }));

    const stats = {
      totalCount: items.length,
      recentUpdates,
    };

    return {
      filteredItems: sorted,
      stats,
    };
  }, [
    items,
    searchQuery,
    sortBy,
    getNamespace,
    getName,
    getDescription,
    getPopularity,
    getLastUpdate,
    getForkOf,
    getDisplayName,
  ]);

  return {
    filteredItems,
    stats,
    sortBy,
    setSortBy,
  };
}
