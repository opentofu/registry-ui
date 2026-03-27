import { SearchResult, SearchResultType } from "./types";

interface SearchProviderResultProps {
  result: SearchResult;
}

export function SearchProviderResult({ result }: SearchProviderResultProps) {
  const namespace = result.addr?.split('/')[0];
  
  return (
    <div className="flex items-start gap-3">
      {namespace && (
        <img 
          src={`https://avatars.githubusercontent.com/${namespace}`} 
          alt={`${namespace} avatar`}
          className="w-8 h-8 rounded ring-1 ring-gray-200 dark:ring-gray-700 flex-shrink-0"
          onError={(e) => {
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
            const fallback = target.nextElementSibling as HTMLElement;
            if (fallback) fallback.style.display = 'flex';
          }}
        />
      )}
      <div 
        className="w-8 h-8 bg-brand-100 dark:bg-brand-900 rounded flex items-center justify-center flex-shrink-0"
        style={{ display: namespace ? 'none' : 'flex' }}
      >
        <span className="text-xs font-semibold text-brand-700 dark:text-brand-300">
          {namespace ? namespace[0].toUpperCase() : 'P'}
        </span>
      </div>
      <div className="flex-grow min-w-0">
        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
          {result.type === SearchResultType.ProviderResource && `Resource: ${result.displayTitle}`}
          {result.type === SearchResultType.ProviderDatasource && `Data source: ${result.displayTitle}`}
          {result.type === SearchResultType.ProviderFunction && `Function: ${result.displayTitle}`}
          {result.type === SearchResultType.Provider && result.addr}
        </div>
        {result.description && (
          <div className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
            {result.description}
          </div>
        )}
      </div>
    </div>
  );
}
