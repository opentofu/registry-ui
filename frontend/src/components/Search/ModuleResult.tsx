import { SearchResult } from "./types";

interface SearchModuleResultProps {
  result: SearchResult;
}

export function SearchModuleResult({ result }: SearchModuleResultProps) {
  return (
    <>
      <div className="text-sm">{result.displayTitle}</div>
      <div className="line-clamp-3 text-xs text-gray-500">
        {result.description}
      </div>
    </>
  );
}
