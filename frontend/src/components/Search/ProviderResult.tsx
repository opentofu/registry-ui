import { SearchResult, SearchResultType } from "./types";

interface SearchProviderResultProps {
  result: SearchResult;
}

export function SearchProviderResult({ result }: SearchProviderResultProps) {
  const isProvider = result.type === SearchResultType.Provider;
  const namespace = result.addr?.split('/')[0];
  
  if (result.type === SearchResultType.ProviderResource) {
    return (
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-8 h-8 bg-gray-100 dark:bg-gray-800 rounded flex items-center justify-center">
          <svg className="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </div>
        <div className="flex-grow min-w-0">
          <div className="text-sm font-medium text-gray-900 dark:text-gray-100">Resource: {result.displayTitle}</div>
          <div className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
            {result.description}
          </div>
        </div>
      </div>
    );
  } else if (result.type === SearchResultType.ProviderDatasource) {
    return (
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-8 h-8 bg-gray-100 dark:bg-gray-800 rounded flex items-center justify-center">
          <svg className="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
          </svg>
        </div>
        <div className="flex-grow min-w-0">
          <div className="text-sm font-medium text-gray-900 dark:text-gray-100">Data source: {result.displayTitle}</div>
          <div className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
            {result.description}
          </div>
        </div>
      </div>
    );
  } else if (result.type === SearchResultType.ProviderFunction) {
    return (
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-8 h-8 bg-gray-100 dark:bg-gray-800 rounded flex items-center justify-center">
          <svg className="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
        </div>
        <div className="flex-grow min-w-0">
          <div className="text-sm font-medium text-gray-900 dark:text-gray-100">Function: {result.displayTitle}</div>
          <div className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
            {result.description}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-start gap-3">
      {isProvider && namespace && (
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
        style={{ display: isProvider && namespace ? 'none' : 'flex' }}
      >
        <span className="text-xs font-semibold text-brand-700 dark:text-brand-300">
          {namespace ? namespace[0].toUpperCase() : 'P'}
        </span>
      </div>
      <div className="flex-grow min-w-0">
        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{result.addr}</div>
        {result.description && (
          <div className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
            {result.description}
          </div>
        )}
      </div>
    </div>
  );
}
