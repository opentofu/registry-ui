import { SearchResult, SearchResultType } from "./types";

interface SearchProviderResultProps {
  result: SearchResult;
}

export function SearchProviderResult({ result }: SearchProviderResultProps) {
  if (result.type === SearchResultType.ProviderResource) {
    return (
      <>
        <div className="text-sm">Resource: {result.title}</div>
        <div className="text-sm">{result.addr}</div>
        <div className="text-xs text-gray-500">{result.description}</div>
      </>
    );
  } else if (result.type === SearchResultType.ProviderDatasource) {
    return (
      <>
        <div className="text-sm">Data source: {result.title}</div>
        <div className="text-sm">{result.addr}</div>
        <div className="text-xs text-gray-500">{result.description}</div>
      </>
    );
  } else if (result.type === SearchResultType.ProviderFunction) {
    return (
      <>
        <div className="text-sm">Function: {result.title}</div>
        <div className="text-sm">{result.addr}</div>
        <div className="text-xs text-gray-500">{result.description}</div>
      </>
    );
  }

  return (
    <>
      <div className="text-sm">{result.addr}</div>
      <div className="text-xs text-gray-500">{result.description}</div>
    </>
  );
}
