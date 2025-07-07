import { SearchResult } from "./types";

interface SearchOtherResultProps {
  result: SearchResult;
}

export function SearchOtherResult({ result }: SearchOtherResultProps) {
  return (
    <div className="flex items-start gap-3">
      <div className="flex-shrink-0 w-8 h-8 bg-gray-100 dark:bg-gray-800 rounded flex items-center justify-center">
        <svg className="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
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
