import { SearchResult, SearchResultType } from "./types";

interface SearchProviderResultProps {
  result: SearchResult;
}

export function SearchProviderResult({ result }: SearchProviderResultProps) {
  const description = (
    <div className="line-clamp-3 text-xs text-gray-500">
      {result.description}
    </div>
  );

  if (result.type === SearchResultType.ProviderResource) {
    return (
      <>
        <div className="text-sm">Resource: {result.displayTitle}</div>
        {description}
      </>
    );
  } else if (result.type === SearchResultType.ProviderDatasource) {
    return (
      <>
        <div className="text-sm">Data source: {result.displayTitle}</div>
        {description}
      </>
    );
  } else if (result.type === SearchResultType.ProviderFunction) {
    return (
      <>
        <div className="text-sm">Function: {result.displayTitle}</div>
        {description}
      </>
    );
  }

  return (
    <>
      <div className="text-sm">{result.addr}</div>
      {description}
    </>
  );
}
