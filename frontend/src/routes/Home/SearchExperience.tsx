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
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";
import { Paragraph } from "@/components/Paragraph";

interface GroupedResults {
  providers: definitions["SearchResultItem"][];
  modules: definitions["SearchResultItem"][];
  resources: definitions["SearchResultItem"][];
  datasources: definitions["SearchResultItem"][];
  functions: definitions["SearchResultItem"][];
}

export function SearchExperience() {
  const [query, setQuery] = useState("");
  const [isSearchActive, setIsSearchActive] = useState(false);
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

    document.addEventListener("keydown", handleSlash);
    return () => document.removeEventListener("keydown", handleSlash);
  }, []);

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
    } else if (e.key === "Escape" && isSearchActive) {
      setQuery("");
      setIsSearchActive(false);
    }
  }, [flatResults, isSearchActive]);

  const handleResultClick = (result: definitions["SearchResultItem"]) => {
    const index = flatResults.findIndex(r => r.id === result.id);
    setSelectedIndex(index);
    setSelectedResult(result);
  };

  const handleSearchInput = (value: string) => {
    setQuery(value);
    if (value.length > 0 && !isSearchActive) {
      setIsSearchActive(true);
    }
  };

  const isLoading = isFetching || query !== debouncedQuery;

  return (
    <>
      <PatternBg />
      <div className={clsx(
        "fixed inset-0 -z-10",
        isSearchActive ? "bg-white/50 dark:bg-blue-950/50" : ""
      )} />
      <Header />
      
      {/* Landing Page Content - Hidden when search is active */}
      <main className={clsx(
        "min-h-screen",
        isSearchActive ? "hidden" : "block"
      )}>
        <div className="container m-auto flex flex-col items-center text-center pt-24">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 340 368"
            className="size-24"
          >
            <path
              fill="#0D1A2B"
              fillRule="evenodd"
              d="M182.26 3.88a25.74 25.74 0 0 0-24.8 0L30.4 73.73l-.31.17-16.74 9.2A25.74 25.74 0 0 0 0 105.66v157.1c0 9.39 5.12 18.03 13.34 22.56l128.28 70.5.5.27 15.34 8.43a25.74 25.74 0 0 0 24.8 0l15.39-8.45.45-.25 128.28-70.5a25.74 25.74 0 0 0 13.34-22.56v-157.1c0-9.4-5.11-18.04-13.34-22.56l-16.67-9.16-.37-.21L182.26 3.88Zm8.17 180.32 118.9-65.35.37-.2 1.42-.79a5.94 5.94 0 0 1 8.8 5.2v122.29a5.94 5.94 0 0 1-8.8 5.2l-1.1-.6-.68-.39-118.91-65.36ZM30.08 118.68l.31.17L149.3 184.2 30.4 249.56l-.67.38-1.12.61a5.94 5.94 0 0 1-8.8-5.2V123.07a5.94 5.94 0 0 1 8.8-5.2l1.47.8Zm269.9-27.5L188.56 29.95a5.94 5.94 0 0 0-8.8 5.2v132.31l120.21-66.07a5.94 5.94 0 0 0 0-10.2Zm-260.2 10.23 120.18 66.05V34.98a5.94 5.94 0 0 0-8.8-5.03L39.78 91.17a5.94 5.94 0 0 0 0 10.24Zm.15 175.92c-4-2.2-4.1-7.85-.31-10.23l120.34-66.15v132.49a5.94 5.94 0 0 1-8.56 5.16L39.93 277.33Zm139.83 55.93v-132.3l120.36 66.15a5.94 5.94 0 0 1-.32 10.22l-111.46 61.26a5.94 5.94 0 0 1-8.58-5.15v-.18Z"
              clipRule="evenodd"
            />
            <path
              fill="#E7C200"
              d="M167 21.23a5.94 5.94 0 0 1 5.72 0L299.8 91.07a5.94 5.94 0 0 1 0 10.42l-127.08 69.84a5.94 5.94 0 0 1-5.72 0L39.93 101.5a5.94 5.94 0 0 1 0-10.42L167 21.23Z"
            />
            <path
              fill="#FFDA18"
              d="M19.8 123.06a5.94 5.94 0 0 1 8.8-5.2l128.28 70.5a5.94 5.94 0 0 1 3.08 5.2v139.7a5.94 5.94 0 0 1-8.8 5.2l-128.28-70.5a5.94 5.94 0 0 1-3.08-5.2v-139.7Z"
            />
            <path
              fill="#fff"
              d="M311.12 117.86a5.94 5.94 0 0 1 8.8 5.2v139.7a5.94 5.94 0 0 1-3.08 5.2l-128.28 70.5a5.94 5.94 0 0 1-8.8-5.2v-139.7a5.94 5.94 0 0 1 3.08-5.2l128.28-70.5Z"
            />
            <path
              fill="#0D1A2B"
              d="m73.68 232.66-.02.2-30.06-15.82v-.2c.7-9.35 8-13.4 16.3-9.03 8.3 4.37 14.48 15.5 13.78 24.85ZM121 259.98l-.02.2-30.07-15.82.02-.2c.7-9.35 7.99-13.4 16.3-9.03 8.3 4.37 14.46 15.5 13.77 24.85Z"
            />
          </svg>
          <h2 className="mt-5 max-w-4xl text-balance text-5xl lg:text-6xl font-bold leading-tight">
            Documentation for OpenTofu Providers and Modules
          </h2>
          <Paragraph className="mb-7 mt-5 text-balance">
            This technology preview contains documentation for a select few
            providers, namespaces, and modules in the OpenTofu registry.
            <br />
            <strong>Note:</strong> the data in this interface may not be up to
            date during the beta phase.
          </Paragraph>
          
          {/* Search Bar for Landing Page */}
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
                onChange={(e) => handleSearchInput(e.target.value)}
                onFocus={() => {
                  if (query.length > 0) {
                    setIsSearchActive(true);
                  }
                }}
                placeholder="Search providers, resources, or modules (Press / to focus)"
                className="w-full h-14 pl-12 pr-4 text-base bg-white border border-gray-200 rounded-xl shadow-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent dark:bg-blue-900 dark:border-gray-700 dark:text-gray-200 dark:placeholder-gray-400"
              />
            </div>
          </div>
        </div>
      </main>

      {/* Search Experience - Shown when search is active */}
      <div className={clsx(
        "mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 pt-24",
        isSearchActive 
          ? "block" 
          : "hidden"
      )}>
        <div className="h-10 bg-gray-200 dark:bg-blue-950 border border-gray-300 dark:border-gray-700 border-b-0 flex items-center px-3 rounded-t">
          <nav className="flex h-10 items-center space-x-2" aria-label="Breadcrumbs">
            <Link
              to="/"
              className="text-gray-700 dark:text-gray-300"
              aria-label="Home"
              onClick={(e) => {
                e.preventDefault();
                setQuery("");
                setIsSearchActive(false);
              }}
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
                  type="text"
                  value={query}
                  onChange={(e) => handleSearchInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Search documentation..."
                  className="w-full h-9 pl-9 pr-9 text-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-md placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent dark:text-gray-200"
                  autoFocus={isSearchActive}
                />
                <button
                  onClick={() => {
                    setQuery("");
                    setIsSearchActive(false);
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

      {!isSearchActive && <Footer />}
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
        "w-full flex items-start gap-3 px-3 py-2 text-left text-sm rounded-md",
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