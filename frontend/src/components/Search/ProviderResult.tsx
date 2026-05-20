import { SearchResult, SearchResultType } from "./types";

interface SearchProviderResultProps {
  result: SearchResult;
}

export function SearchProviderResult({ result }: SearchProviderResultProps) {
  const namespace = result.addr?.split("/")[0];

  return (
    <div className="flex items-start gap-3">
      {namespace && (
        <img
          src={`https://avatars.githubusercontent.com/${namespace}`}
          alt={`${namespace} avatar`}
          className="h-8 w-8 flex-shrink-0 rounded ring-1 ring-gray-200 dark:ring-gray-700"
          onError={(e) => {
            const target = e.target as HTMLImageElement;
            target.style.display = "none";
            const fallback = target.nextElementSibling as HTMLElement;
            if (fallback) fallback.style.display = "flex";
          }}
        />
      )}
      <div
        className="bg-brand-100 dark:bg-brand-900 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded"
        style={{ display: namespace ? "none" : "flex" }}
      >
        <span className="text-brand-700 dark:text-brand-300 text-xs font-semibold">
          {namespace ? namespace[0].toUpperCase() : "P"}
        </span>
      </div>
      <div className="min-w-0 flex-grow">
        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
          {result.type === SearchResultType.ProviderResource &&
            `Resource: ${result.displayTitle}`}
          {result.type === SearchResultType.ProviderDatasource &&
            `Data source: ${result.displayTitle}`}
          {result.type === SearchResultType.ProviderFunction &&
            `Function: ${result.displayTitle}`}
          {result.type === SearchResultType.Provider && result.addr}
        </div>
        {result.description && (
          <div className="mt-0.5 line-clamp-2 text-xs text-gray-500 dark:text-gray-400">
            {result.description}
          </div>
        )}
      </div>
    </div>
  );
}
