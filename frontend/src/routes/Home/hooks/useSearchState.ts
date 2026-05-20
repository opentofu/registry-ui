import { useState } from "react";
import { components } from "@/api";

export interface GroupedResults {
  providers: components["schemas"]["SearchResultItem"][];
  modules: components["schemas"]["SearchResultItem"][];
  resources: components["schemas"]["SearchResultItem"][];
  datasources: components["schemas"]["SearchResultItem"][];
  functions: components["schemas"]["SearchResultItem"][];
}

export function useSearchState() {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [selectedResult, setSelectedResult] = useState<
    components["schemas"]["SearchResultItem"] | null
  >(null);

  const handleSearchInput = (value: string) => {
    setQuery(value);
    // Clear selection when search changes
    if (selectedResult && value !== query) {
      setSelectedResult(null);
      setSelectedIndex(0);
    }
  };

  const handleResultClick = (
    result: components["schemas"]["SearchResultItem"],
    flatResults: components["schemas"]["SearchResultItem"][],
  ) => {
    const index = flatResults.findIndex((r) => r.id === result.id);
    setSelectedIndex(index);
    setSelectedResult(result);
  };

  const handleClearSearch = () => {
    setQuery("");
    setSelectedResult(null);
    setSelectedIndex(0);
  };

  const handleHomeClick = () => {
    setQuery("");
    setSelectedResult(null);
    setSelectedIndex(0);
  };

  return {
    query,
    setQuery,
    selectedIndex,
    setSelectedIndex,
    selectedResult,
    setSelectedResult,
    handleSearchInput,
    handleResultClick,
    handleClearSearch,
    handleHomeClick,
  };
}
