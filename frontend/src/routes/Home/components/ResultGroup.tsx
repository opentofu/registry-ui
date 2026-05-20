import { components } from "@/api";
import { ResultItem } from "./ResultItem";

interface ResultGroupProps {
  title: string;
  results: components["schemas"]["SearchResultItem"][];
  selectedResult: components["schemas"]["SearchResultItem"] | null;
  onResultClick: (result: components["schemas"]["SearchResultItem"]) => void;
}

export function ResultGroup({
  title,
  results,
  selectedResult,
  onResultClick,
}: ResultGroupProps) {
  if (results.length === 0) {
    return null;
  }

  return (
    <div className="mb-2">
      <h3 className="px-4 py-2 text-xs font-semibold tracking-wider text-gray-500 uppercase dark:text-gray-400">
        {title}
      </h3>
      <div className="space-y-1 px-2">
        {results.map((result) => (
          <ResultItem
            key={result.id}
            result={result}
            isSelected={selectedResult?.id === result.id}
            onClick={() => onResultClick(result)}
          />
        ))}
      </div>
    </div>
  );
}
