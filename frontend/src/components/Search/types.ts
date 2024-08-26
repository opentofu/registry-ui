export enum SearchResultType {
  Provider = "provider",
  Module = "module",
  ProviderResource = "provider/resource",
  ProviderDatasource = "provider/datasource",
  ProviderFunction = "provider/function",
  Other = "other",
}

export interface ApiSearchResult {
  id: string;
  type: SearchResultType;
  addr: string;
  title: string;
  description: string;
  version: string;
  link_variables: {
    name: string;
    version: string;
    namespace: string;
    target_system?: string;
  };
}

export interface SearchResult {
  id: string;
  addr: string;
  title: string;
  description: string;
  link: string;
  type: SearchResultType;
}
