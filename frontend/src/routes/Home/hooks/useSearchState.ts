import { useState, useMemo } from "react";
import { definitions } from "@/api";

export interface GroupedResults {
  providers: definitions["SearchResultItem"][];
  modules: definitions["SearchResultItem"][];
  resources: definitions["SearchResultItem"][];
  datasources: definitions["SearchResultItem"][];
  functions: definitions["SearchResultItem"][];
}

export function useSearchState() {
  const [query, setQuery] = useState("");
  const [isSearchActive, setIsSearchActive] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [selectedResult, setSelectedResult] = useState<
    definitions["SearchResultItem"] | null
  >(null);


  const handleSearchInput = (value: string) => {
    setQuery(value);
    if (value.length > 0 && !isSearchActive) {
      setIsSearchActive(true);
    }
  };

  const handleResultClick = (result: definitions["SearchResultItem"], flatResults: definitions["SearchResultItem"][]) => {
    const index = flatResults.findIndex((r) => r.id === result.id);
    setSelectedIndex(index);
    setSelectedResult(result);
  };

  const handleClearSearch = () => {
    setQuery("");
    setIsSearchActive(false);
  };

  const handleHomeClick = () => {
    setQuery("");
    setIsSearchActive(false);
  };

  return {
    query,
    setQuery,
    isSearchActive,
    setIsSearchActive,
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