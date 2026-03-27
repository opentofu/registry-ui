import { definitions } from "@/api";
import { SearchInput } from "./SearchInput";
import { ResultGroup } from "./ResultGroup";

interface GroupedResults {
  providers: definitions["SearchResultItem"][];
  modules: definitions["SearchResultItem"][];
  resources: definitions["SearchResultItem"][];
  datasources: definitions["SearchResultItem"][];
  functions: definitions["SearchResultItem"][];
}

interface SearchResultsProps {
  query: string;
  onQueryChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
  onClear: () => void;
  isLoading: boolean;
  groupedResults: GroupedResults | null;
  flatResults: definitions["SearchResultItem"][];
  selectedResult: definitions["SearchResultItem"] | null;
  onResultClick: (result: definitions["SearchResultItem"]) => void;
  resultsContainerRef: React.RefObject<HTMLDivElement>;
  isSearchActive: boolean;
}

export function SearchResults({
  query,
  onQueryChange,
  onKeyDown,
  onClear,
  isLoading,
  groupedResults,
  flatResults,
  selectedResult,
  onResultClick,
  resultsContainerRef,
  isSearchActive,
}: SearchResultsProps) {
  return (
    <aside className="custom-scrollbar sticky top-0 flex max-h-screen w-1/5 min-w-80 shrink-0 flex-col overflow-y-auto border-r border-gray-200 bg-gray-100 pt-2 dark:border-gray-800 dark:bg-blue-900">
      {/* Search Input */}
      <div className="border-b border-gray-200 p-4 dark:border-gray-700">
        <SearchInput
          value={query}
          onChange={onQueryChange}
          onKeyDown={onKeyDown}
          placeholder="Search documentation..."
          size="small"
          showClearButton={true}
          onClear={onClear}
          autoFocus={isSearchActive}
        />
        <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
          Use ↑↓ to navigate • Enter to open • ESC to close
        </div>
      </div>

      {/* Search Results */}
      <div ref={resultsContainerRef} className="flex-1 overflow-y-auto">
        {isLoading && (
          <div className="flex justify-center py-8">
            <div className="border-brand-500 h-6 w-6 animate-spin rounded-full border-b-2"></div>
          </div>
        )}

        {!isLoading && flatResults.length === 0 && query && (
          <div className="px-4 py-8 text-center">
            <p className="text-sm text-gray-500 dark:text-gray-400">
              No results found for "{query}"
            </p>
          </div>
        )}

        {!isLoading && groupedResults && flatResults.length > 0 && (
          <div className="py-2">
            <ResultGroup
              title="Providers"
              results={groupedResults.providers}
              selectedResult={selectedResult}
              onResultClick={onResultClick}
            />
            <ResultGroup
              title="Resources"
              results={groupedResults.resources}
              selectedResult={selectedResult}
              onResultClick={onResultClick}
            />
            <ResultGroup
              title="Data Sources"
              results={groupedResults.datasources}
              selectedResult={selectedResult}
              onResultClick={onResultClick}
            />
            <ResultGroup
              title="Functions"
              results={groupedResults.functions}
              selectedResult={selectedResult}
              onResultClick={onResultClick}
            />
            <ResultGroup
              title="Modules"
              results={groupedResults.modules}
              selectedResult={selectedResult}
              onResultClick={onResultClick}
            />
          </div>
        )}
      </div>
    </aside>
  );
}
