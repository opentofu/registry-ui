import { useEffect, useRef, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import clsx from "clsx";

import { Footer } from "@/components/Footer";
import { Header } from "@/components/Header";
import PatternBg from "@/components/PatternBg";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { getSearchQuery } from "@/q";

import { DocumentationPreview } from "./DocumentationPreview";
import { LandingPage } from "./components/LandingPage";
import { SearchBreadcrumbs } from "./components/SearchBreadcrumbs";
import { SearchResults } from "./components/SearchResults";
import { useSearchState, GroupedResults } from "./hooks/useSearchState";
import { useSearchKeyboard } from "./hooks/useSearchKeyboard";

export function SearchExperience() {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const resultsContainerRef = useRef<HTMLDivElement | null>(null);

  // Initialize search state
  const {
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
  } = useSearchState();

  // Get search data based on user query
  const debouncedQuery = useDebouncedValue(query, 250);
  const { data: searchData, isFetching } = useQuery(
    getSearchQuery(debouncedQuery),
  );

  // Process the search results
  const { groupedResults, flatResults } = useMemo(() => {
    if (!searchData || searchData.length === 0)
      return { groupedResults: null, flatResults: [] };

    const groups: GroupedResults = {
      providers: [],
      modules: [],
      resources: [],
      datasources: [],
      functions: [],
    };

    const limited = searchData.slice(0, 30);
    limited.forEach((result) => {
      if (result.type === "provider") {
        groups.providers.push(result);
      } else if (result.type === "module") {
        groups.modules.push(result);
      } else if (result.type === "provider/resource") {
        groups.resources.push(result);
      } else if (result.type === "provider/datasource") {
        groups.datasources.push(result);
      } else if (result.type === "provider/function") {
        groups.functions.push(result);
      }
    });

    return { groupedResults: groups, flatResults: limited };
  }, [searchData]);

  // Setup keyboard navigation
  const { handleKeyDown } = useSearchKeyboard({
    flatResults,
    selectedIndex,
    setSelectedIndex,
    setSelectedResult,
    resultsContainerRef,
    isSearchActive,
    setQuery,
    setIsSearchActive,
  });

  // Auto-select first result when results change
  useEffect(() => {
    if (flatResults.length > 0 && selectedIndex === 0) {
      setSelectedResult(flatResults[0]);
    }
  }, [flatResults, selectedIndex, setSelectedResult]);

  // Setup slash key to focus search
  useEffect(() => {
    const handleSlash = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement;
      const isInputOrTextarea =
        target instanceof HTMLInputElement ||
        target instanceof HTMLTextAreaElement;

      if (
        event.key === "/" &&
        !isInputOrTextarea &&
        target !== inputRef.current
      ) {
        event.preventDefault();
        inputRef.current?.focus();
      }
    };

    document.addEventListener("keydown", handleSlash);
    return () => document.removeEventListener("keydown", handleSlash);
  }, []);

  const isLoading = isFetching || query !== debouncedQuery;

  return (
    <>
      <PatternBg />
      <div
        className={clsx(
          "fixed inset-0 -z-10",
          isSearchActive ? "bg-white/50 dark:bg-blue-950/50" : "",
        )}
      />
      <Header />

      {/* Landing Page Content - Hidden when search is active */}
      {!isSearchActive && (
        <LandingPage
          query={query}
          onQueryChange={handleSearchInput}
          onSearchFocus={() => setIsSearchActive(true)}
          inputRef={inputRef}
        />
      )}

      {/* Search Experience - Shown when search is active */}
      {isSearchActive && (
        <div className="mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 pt-24">
          <SearchBreadcrumbs onHomeClick={handleHomeClick} />
          <div className="flex flex-1 divide-x divide-gray-200 border border-t-0 border-gray-300 dark:divide-gray-800 dark:border-gray-700">
            <SearchResults
              query={query}
              onQueryChange={handleSearchInput}
              onKeyDown={handleKeyDown}
              onClear={handleClearSearch}
              isLoading={isLoading}
              groupedResults={groupedResults}
              flatResults={flatResults}
              selectedResult={selectedResult}
              onResultClick={(result) => handleResultClick(result, flatResults)}
              resultsContainerRef={resultsContainerRef}
              isSearchActive={isSearchActive}
            />

            {/* Right Side - Documentation Preview */}
            <main className="min-w-0 flex-1 bg-gray-100 dark:bg-blue-900">
              <div className="mt-8">
                {selectedResult ? (
                  <DocumentationPreview result={selectedResult} />
                ) : (
                  <div className="mt-8 flex h-full items-center justify-center p-8 text-gray-400 dark:text-gray-600">
                    <div className="text-center">
                      <p className="text-sm">
                        Select a result to preview documentation
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </main>
          </div>
        </div>
      )}

      {!isSearchActive && <Footer />}
    </>
  );
}
