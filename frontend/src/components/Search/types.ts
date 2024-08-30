export enum SearchResultType {
  Provider = "provider",
  Module = "module",
  ProviderResource = "provider/resource",
  ProviderDatasource = "provider/datasource",
  ProviderFunction = "provider/function",
  Other = "other",
}

export interface SearchResult {
  id: string;
  addr: string;
  title: string;
  description: string;
  link: string;
  type: SearchResultType;
}
