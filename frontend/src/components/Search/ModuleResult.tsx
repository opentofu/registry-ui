import { SearchResult } from "./types";

interface SearchModuleResultProps {
  result: SearchResult;
}

export function SearchModuleResult({ result }: SearchModuleResultProps) {
  const namespace = result.displayTitle?.split('/')[0];
  
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
        className="w-8 h-8 bg-purple-100 dark:bg-purple-900 rounded flex items-center justify-center flex-shrink-0"
        style={{ display: namespace ? 'none' : 'flex' }}
      >
        <svg className="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
        </svg>
      </div>
      <div className="flex-grow min-w-0">
        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{result.displayTitle}</div>
        {result.description && (
          <div className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
            {result.description}
          </div>
        )}
      </div>
    </div>
  );
}
