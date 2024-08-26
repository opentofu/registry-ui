import { SearchResult } from "./types";

interface SearchModuleResultProps {
  result: SearchResult;
}

export function SearchModuleResult({ result }: SearchModuleResultProps) {
  return (
    <>
      <div className="text-sm">{result.addr}</div>
      <div className="text-xs text-gray-500">{result.description}</div>
    </>
  );
}
