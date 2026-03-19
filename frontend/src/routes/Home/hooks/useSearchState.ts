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
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [selectedResult, setSelectedResult] = useState<
    definitions["SearchResultItem"] | null
  >(null);

  const handleSearchInput = (value: string) => {
    setQuery(value);
    // Clear selection when search changes
    if (selectedResult && value !== query) {
      setSelectedResult(null);
      setSelectedIndex(0);
    }
  };

  const handleResultClick = (result: definitions["SearchResultItem"], flatResults: definitions["SearchResultItem"][]) => {
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