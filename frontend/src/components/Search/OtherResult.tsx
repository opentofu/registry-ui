import { SearchResult } from "./types";

interface SearchOtherResultProps {
  result: SearchResult;
}

export function SearchOtherResult({ result }: SearchOtherResultProps) {
  return (
    <>
      <div className="text-sm">{result.addr}</div>
      <div className="line-clamp-3 text-xs text-gray-500">
        {result.description}
      </div>
    </>
  );
}
