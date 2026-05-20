import { SearchResult } from "./types";

interface SearchOtherResultProps {
  result: SearchResult;
}

export function SearchOtherResult({ result }: SearchOtherResultProps) {
  return (
    <div className="flex items-start gap-3">
      <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded bg-gray-100 dark:bg-gray-800">
        <svg
          className="h-4 w-4 text-gray-600 dark:text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
      </div>
      <div className="min-w-0 flex-grow">
        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
          {result.addr}
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
