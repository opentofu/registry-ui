import { useState, useEffect, useMemo, useRef, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import clsx from "clsx";
import { definitions } from "@/api";
import { getSearchQuery } from "@/q";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { Icon } from "@/components/Icon";
import { search as searchIcon } from "@/icons/search";
import { DocumentationPreview } from "./DocumentationPreview";
import PatternBg from "@/components/PatternBg";
import { Link } from "react-router";
import { home } from "@/icons/home";
import { chevron } from "@/icons/chevron";

interface SearchExperienceProps {
  onSearchStateChange: (isActive: boolean) => void;
  fullView?: boolean;
}

interface GroupedResults {
  providers: definitions["SearchResultItem"][];
  modules: definitions["SearchResultItem"][];
  resources: definitions["SearchResultItem"][];
  datasources: definitions["SearchResultItem"][];
  functions: definitions["SearchResultItem"][];
}

export function SearchExperience({ onSearchStateChange, fullView = false }: SearchExperienceProps) {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [selectedResult, setSelectedResult] = useState<definitions["SearchResultItem"] | null>(null);
  const debouncedQuery = useDebouncedValue(query, 250);
  const { data, isFetching } = useQuery(getSearchQuery(debouncedQuery));
  const inputRef = useRef<HTMLInputElement | null>(null);
  const resultsContainerRef = useRef<HTMLDivElement | null>(null);

  const { groupedResults, flatResults } = useMemo(() => {
    if (!data || data.length === 0) return { groupedResults: null, flatResults: [] };
    
    const groups: GroupedResults = {
      providers: [],
      modules: [],
      resources: [],
      datasources: [],
      functions: [],
    };

    const limited = data.slice(0, 30);
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
  }, [data]);

  useEffect(() => {
    if (!fullView) {
      onSearchStateChange(false);
    }
  }, [fullView, onSearchStateChange]);

  useEffect(() => {
    // Auto-select first result when results change
    if (flatResults.length > 0 && selectedIndex === 0) {
      setSelectedResult(flatResults[0]);
    }
  }, [flatResults, selectedIndex]);

  useEffect(() => {
    const handleSlash = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement;
      const isInputOrTextarea = target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement;
      
      if (event.key === "/" && !isInputOrTextarea && target !== inputRef.current) {
        event.preventDefault();
        inputRef.current?.focus();
      }
    };

    if (!fullView) {
      document.addEventListener("keydown", handleSlash);
      return () => document.removeEventListener("keydown", handleSlash);
    }
  }, [fullView]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex(prev => {
        const newIndex = Math.min(prev + 1, flatResults.length - 1);
        setSelectedResult(flatResults[newIndex]);
        // Scroll to selected item
        const container = resultsContainerRef.current;
        const items = container?.querySelectorAll('[data-result-item]');
        if (items && items[newIndex]) {
          items[newIndex].scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        }
        return newIndex;
      });
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex(prev => {
        const newIndex = Math.max(prev - 1, 0);
        setSelectedResult(flatResults[newIndex]);
        // Scroll to selected item
        const container = resultsContainerRef.current;
        const items = container?.querySelectorAll('[data-result-item]');
        if (items && items[newIndex]) {
          items[newIndex].scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        }
        return newIndex;
      });
    } else if (e.key === "Escape" && fullView) {
      setQuery("");
      onSearchStateChange(false);
    }
  }, [flatResults, fullView, onSearchStateChange]);

  const handleResultClick = (result: definitions["SearchResultItem"]) => {
    const index = flatResults.findIndex(r => r.id === result.id);
    setSelectedIndex(index);
    setSelectedResult(result);
  };

  const isLoading = isFetching || query !== debouncedQuery;

  if (!fullView) {
    // Simple search bar for landing page
    return (
      <div className="w-full max-w-xl">
        <div className="relative">
          <Icon
            path={searchIcon}
            className="absolute left-4 top-1/2 -translate-y-1/2 size-5 text-gray-400"
          />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              if (e.target.value.length > 0) {
                onSearchStateChange(true);
              }
            }}
            onFocus={() => {
              if (query.length > 0) {
                onSearchStateChange(true);
              }
            }}
            placeholder="Search providers, resources, or modules (Press / to focus)"
            className="w-full h-14 pl-12 pr-4 text-base bg-white border border-gray-200 rounded-xl shadow-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent dark:bg-blue-900 dark:border-gray-700 dark:text-gray-200 dark:placeholder-gray-400"
          />
        </div>
      </div>
    );
  }

  // Full split-view search experience with Provider-style layout
  return (
    <>
      <PatternBg />
      <div className="fixed inset-0 -z-10 bg-white/50 dark:bg-blue-950/50" />
      <div className="mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 pt-24">
        <div className="h-10 bg-gray-200 dark:bg-blue-950 border border-gray-300 dark:border-gray-700 border-b-0 flex items-center px-3 rounded-t">
          <nav className="flex h-10 items-center space-x-2" aria-label="Breadcrumbs">
            <Link
              to="/"
              className="text-gray-700 dark:text-gray-300"
              aria-label="Home"
            >
              <Icon path={home} className="size-5" />
            </Link>
            <span>
              <Icon
                path={chevron}
                className="size-4 text-gray-400 dark:text-gray-600"
              />
            </span>
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Search
            </span>
          </nav>
        </div>
        <div className="flex flex-1 divide-x divide-gray-200 dark:divide-gray-800 border border-gray-300 dark:border-gray-700 border-t-0">
          {/* Left Sidebar - Search Results */}
          <aside className="sticky top-0 flex max-h-screen w-1/5 min-w-80 shrink-0 flex-col overflow-y-auto custom-scrollbar bg-gray-100 dark:bg-blue-900 border-r border-gray-200 dark:border-gray-800 pt-2">
            {/* Search Input */}
            <div className="p-4 border-b border-gray-200 dark:border-gray-700">
              <div className="relative">
                <Icon
                  path={searchIcon}
                  className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-gray-400"
                />
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Search documentation..."
                  className="w-full h-9 pl-9 pr-9 text-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-md placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent dark:text-gray-200"
                  autoFocus
                />
                <button
                  onClick={() => {
                    setQuery("");
                    onSearchStateChange(false);
                  }}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                Use ↑↓ to navigate • ESC to close
              </div>
            </div>

            {/* Search Results */}
            <div ref={resultsContainerRef} className="flex-1 overflow-y-auto">
              {isLoading && (
                <div className="flex justify-center py-8">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-brand-500"></div>
                </div>
              )}

              {!isLoading && flatResults.length === 0 && query && (
                <div className="text-center py-8 px-4">
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    No results found for "{query}"
                  </p>
                </div>
              )}

              {!isLoading && groupedResults && flatResults.length > 0 && (
                <div className="py-2">
                  {/* Providers */}
                  {groupedResults.providers.length > 0 && (
                    <div className="mb-2">
                      <h3 className="px-4 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Providers
                      </h3>
                      <div className="px-2 space-y-1">
                        {groupedResults.providers.map((result) => (
                          <ResultItem
                            key={result.id}
                            result={result}
                            isSelected={selectedResult?.id === result.id}
                            onClick={() => handleResultClick(result)}
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Resources */}
                  {groupedResults.resources.length > 0 && (
                    <div className="mb-2">
                      <h3 className="px-4 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Resources
                      </h3>
                      <div className="px-2 space-y-1">
                        {groupedResults.resources.map((result) => (
                          <ResultItem
                            key={result.id}
                            result={result}
                            isSelected={selectedResult?.id === result.id}
                            onClick={() => handleResultClick(result)}
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Data Sources */}
                  {groupedResults.datasources.length > 0 && (
                    <div className="mb-2">
                      <h3 className="px-4 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Data Sources
                      </h3>
                      <div className="px-2 space-y-1">
                        {groupedResults.datasources.map((result) => (
                          <ResultItem
                            key={result.id}
                            result={result}
                            isSelected={selectedResult?.id === result.id}
                            onClick={() => handleResultClick(result)}
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Functions */}
                  {groupedResults.functions.length > 0 && (
                    <div className="mb-2">
                      <h3 className="px-4 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Functions
                      </h3>
                      <div className="px-2 space-y-1">
                        {groupedResults.functions.map((result) => (
                          <ResultItem
                            key={result.id}
                            result={result}
                            isSelected={selectedResult?.id === result.id}
                            onClick={() => handleResultClick(result)}
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Modules */}
                  {groupedResults.modules.length > 0 && (
                    <div className="mb-2">
                      <h3 className="px-4 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Modules
                      </h3>
                      <div className="px-2 space-y-1">
                        {groupedResults.modules.map((result) => (
                          <ResultItem
                            key={result.id}
                            result={result}
                            isSelected={selectedResult?.id === result.id}
                            onClick={() => handleResultClick(result)}
                          />
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </aside>

          {/* Right Side - Documentation Preview */}
          <main className="min-w-0 flex-1 bg-gray-100 dark:bg-blue-900">
            <div className="mt-8">
              {selectedResult ? (
                <DocumentationPreview result={selectedResult} />
              ) : (
                <div className="h-full flex items-center justify-center text-gray-400 dark:text-gray-600 p-8 mt-8">
                  <div className="text-center">
                    <svg className="w-16 h-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <p className="text-sm">Select a result to preview documentation</p>
                  </div>
                </div>
              )}
            </div>
          </main>
        </div>
      </div>
    </>
  );
}

// Separate component for result items
interface ResultItemProps {
  result: definitions["SearchResultItem"];
  isSelected: boolean;
  onClick: () => void;
}

function ResultItem({ result, isSelected, onClick }: ResultItemProps) {
  return (
    <button
      data-result-item
      onClick={onClick}
      className={clsx(
        "w-full flex items-start gap-3 px-3 py-2 text-left text-sm rounded-md transition-all duration-150",
        isSelected 
          ? "bg-brand-500/10 text-brand-700 dark:bg-brand-500/20 dark:text-brand-400" 
          : "text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white"
      )}
    >
      <div className="flex-shrink-0 mt-0.5">
        {result.type === "provider" ? (
          <img 
            src={`https://avatars.githubusercontent.com/${result.link_variables.namespace}`} 
            alt=""
            className="w-5 h-5 rounded"
            loading="lazy"
            onError={(e) => {
              e.currentTarget.src = '/favicon.ico';
            }}
          />
        ) : (
          <div className={clsx(
            "w-5 h-5 rounded flex items-center justify-center",
            result.type === "provider/resource" && "bg-orange-100 dark:bg-orange-900/30",
            result.type === "provider/datasource" && "bg-blue-100 dark:bg-blue-900/30",
            result.type === "provider/function" && "bg-purple-100 dark:bg-purple-900/30",
            result.type === "module" && "bg-gradient-to-br from-brand-400 to-brand-600"
          )}>
            <svg className={clsx(
              "w-3 h-3",
              result.type === "provider/resource" && "text-orange-600 dark:text-orange-400",
              result.type === "provider/datasource" && "text-blue-600 dark:text-blue-400",
              result.type === "provider/function" && "text-purple-600 dark:text-purple-400",
              result.type === "module" && "text-white"
            )} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {result.type === "provider/resource" && (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              )}
              {result.type === "provider/datasource" && (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
              )}
              {result.type === "provider/function" && (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              )}
              {result.type === "module" && (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              )}
            </svg>
          </div>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className={clsx(
          "break-all",
          isSelected && "font-medium"
        )}>
          {result.link_variables.namespace}/{result.link_variables.name}
          {result.type !== "provider" && result.type !== "module" && (
            <span className="ml-1 text-gray-500 dark:text-gray-400">
              → {result.link_variables.id}
            </span>
          )}
        </div>
        {result.description && (
          <div className="mt-0.5 text-xs text-gray-600 dark:text-gray-400 line-clamp-1">
            {result.description}
          </div>
        )}
      </div>
    </button>
  );
}