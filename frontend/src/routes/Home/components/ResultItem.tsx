import clsx from "clsx";
import { definitions } from "@/api";

interface ResultItemProps {
  result: definitions["SearchResultItem"];
  isSelected: boolean;
  onClick: () => void;
}

export function ResultItem({ result, isSelected, onClick }: ResultItemProps) {
  return (
    <button
      data-result-item
      onClick={onClick}
      className={clsx(
        "flex w-full items-start gap-3 rounded-md px-3 py-2 text-left text-sm",
        isSelected
          ? "bg-brand-500/10 text-brand-700 dark:bg-brand-500/20 dark:text-brand-400"
          : "text-gray-700 hover:bg-gray-200 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white",
      )}
    >
      <div className="mt-0.5 flex-shrink-0">
        <img
          src={`https://avatars.githubusercontent.com/${result.link_variables.namespace}`}
          alt=""
          className="h-5 w-5 rounded"
          loading="lazy"
          onError={(e) => {
            e.currentTarget.src = "/favicon.ico";
          }}
        />
      </div>
      <div className="min-w-0 flex-1">
        <div className={clsx("break-all", isSelected && "font-medium")}>
          {result.link_variables.namespace}/{result.link_variables.name}
          {result.type !== "provider" && result.type !== "module" && (
            <span className="ml-1 text-gray-500 dark:text-gray-400">
              → {result.link_variables.id}
            </span>
          )}
        </div>
        {result.description && (
          <div className="mt-0.5 line-clamp-1 text-xs text-gray-600 dark:text-gray-400">
            {result.description}
          </div>
        )}
      </div>
    </button>
  );
}