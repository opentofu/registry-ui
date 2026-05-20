import { SearchResult } from "./types";

interface SearchModuleResultProps {
  result: SearchResult;
}

export function SearchModuleResult({ result }: SearchModuleResultProps) {
  const namespace = result.displayTitle?.split("/")[0];

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
        className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded bg-purple-100 dark:bg-purple-900"
        style={{ display: namespace ? "none" : "flex" }}
      >
        <svg
          className="h-4 w-4 text-purple-600 dark:text-purple-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
          />
        </svg>
      </div>
      <div className="min-w-0 flex-grow">
        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
          {result.displayTitle}
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
