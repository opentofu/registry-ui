import {
  Combobox,
  ComboboxInput,
  ComboboxOption,
  ComboboxOptions,
} from "@headlessui/react";
import { SearchResult, SearchResultType } from "./types";
import { useEffect, useMemo, useRef, useState } from "react";

import { Icon } from "../Icon";
import { Paragraph } from "../Paragraph";
import { SearchModuleResult } from "./ModuleResult";
import { SearchOtherResult } from "./OtherResult";
import { SearchProviderResult } from "./ProviderResult";
import clsx from "clsx";
import { definitions } from "@/api";
import { getSearchQuery } from "../../q";
import { search } from "../../icons/search";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { useNavigate } from "react-router";
import { useQuery } from "@tanstack/react-query";

function getSearchResultType(value: string) {
  switch (value) {
    case "provider":
      return SearchResultType.Provider;
    case "module":
      return SearchResultType.Module;
    case "provider/resource":
      return SearchResultType.ProviderResource;
    case "provider/datasource":
      return SearchResultType.ProviderDatasource;
    case "provider/function":
      return SearchResultType.ProviderFunction;
    default:
      return SearchResultType.Other;
  }
}

function getSearchResultTypeOrder(type: SearchResultType) {
  switch (type) {
    case SearchResultType.Provider:
    case SearchResultType.ProviderResource:
    case SearchResultType.ProviderDatasource:
    case SearchResultType.ProviderFunction:
      return 0;
    case SearchResultType.Module:
      return 1;
    default:
      return 2;
  }
}

function getSearchResultTypeLabel(type: SearchResultType) {
  switch (type) {
    case SearchResultType.Module:
      return "Modules";
    case SearchResultType.ProviderResource:
    case SearchResultType.ProviderDatasource:
    case SearchResultType.ProviderFunction:
    case SearchResultType.Provider:
      return "Providers";
    case SearchResultType.Other:
      return "Other";
  }
}

function getSearchResultTypeLink(
  type: SearchResultType,
  result: definitions["SearchResultItem"],
) {
  switch (type) {
    case SearchResultType.Module:
      return `/module/${result.link_variables.namespace}/${result.link_variables.name}/${result.link_variables.target_system}/${result.link_variables.version}`;
    case SearchResultType.Provider:
      return `/provider/${result.link_variables.namespace}/${result.link_variables.name}/${result.link_variables.version}`;
    case SearchResultType.ProviderResource:
      return `/provider/${result.link_variables.namespace}/${result.link_variables.name}/${result.link_variables.version}/docs/resources/${result.link_variables.id}`;
    case SearchResultType.ProviderDatasource:
      return `/provider/${result.link_variables.namespace}/${result.link_variables.name}/${result.link_variables.version}/docs/datasources/${result.link_variables.id}`;
    case SearchResultType.ProviderFunction:
      return `/provider/${result.link_variables.namespace}/${result.link_variables.name}/${result.link_variables.version}/docs/functions/${result.link_variables.id}`;
    default:
      return "";
  }
}

function getSearchResultDisplayTitle(
  type: SearchResultType,
  result: definitions["SearchResultItem"],
) {
  switch (type) {
    case SearchResultType.Module:
      return `${result.link_variables.namespace}/${result.link_variables.name}`;
    case SearchResultType.Provider:
      return `${result.link_variables.namespace}/${result.link_variables.name}`;
    case SearchResultType.ProviderResource:
    case SearchResultType.ProviderDatasource:
    case SearchResultType.ProviderFunction:
      return `${result.link_variables.namespace}/${result.link_variables.name} - ${result.link_variables.id}`;
    default:
      return result.title;
  }
}

type Results = Array<{
  label: string;
  type: SearchResultType;
  results: SearchResult[];
}>;

type SearchProps = {
  size?: "small" | "large";
  placeholder?: string;
};

export function Search({
  size = "small",
  placeholder = "Search resources (Press / to focus)",
}: SearchProps) {
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebouncedValue(query, 250);
  const { data, isFetching } = useQuery(getSearchQuery(debouncedQuery));

  const inputRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();

  const filtered = useMemo(() => {
    if (!data) {
      return [];
    }

    const results: Results = [];

    for (let i = 0; i < data.length; i++) {
      if (i >= 10) {
        break;
      }

      const result = data[i];
      const type = getSearchResultType(result.type);
      const order = getSearchResultTypeOrder(type);
      const link = getSearchResultTypeLink(type, result);
      const displayTitle = getSearchResultDisplayTitle(type, result);

      if (!results[order]) {
        results[order] = {
          type,
          label: getSearchResultTypeLabel(type),
          results: [],
        };
      }

      results[order].results.push({
        id: result.id,
        title: result.title,
        addr: result.addr,
        description: result.description,
        link,
        type,
        displayTitle,
      });
    }

    return results;
  }, [data]);

  const onChange = (value: string) => {
    setQuery(value);
  };

  useEffect(() => {
    const handleSlash = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement;
      const isInputOrTextarea = target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement;
      
      if (event.key === "/" && !isInputOrTextarea && target !== inputRef.current) {
        event.preventDefault();
        inputRef.current?.focus();
      }
    };

    document.addEventListener("keydown", handleSlash);

    return () => {
      document.removeEventListener("keydown", handleSlash);
    };
  }, []);

  const canShowLoadingInfo = isFetching || query !== debouncedQuery;
  const canShowNoResultsInfo = filtered.length === 0 && !canShowLoadingInfo;
  const canShowResults = !canShowLoadingInfo && !canShowNoResultsInfo;

  const onKeyDown = (
    event: React.KeyboardEvent<HTMLInputElement>,
    cannotPressEnter: boolean,
  ) => {
    if (event.code === "Enter" && cannotPressEnter) {
      event.preventDefault();
    }
  };

  return (
    <Combobox
      onChange={(v: SearchResult) => {
        if (!v) {
          return;
        }
        navigate(v.link);
      }}
    >
      <div
        className={clsx(
          "relative",
          size === "small" ? "w-96" : "w-full max-w-xl",
        )}
      >
        <Icon
          path={search}
          className={clsx(
            "absolute z-10 text-gray-400",
            size === "small" ? "left-3 top-2.5 size-4" : "left-4 top-3.5 size-5",
          )}
        />
        <ComboboxInput
          ref={inputRef}
          value={query}
          onChange={(event) => onChange(event.target.value)}
          onKeyDown={(event) => onKeyDown(event, canShowLoadingInfo)}
          placeholder={placeholder}
          className={clsx(
            "relative block w-full appearance-none rounded-xl bg-white border border-gray-200 text-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 focus:border-transparent dark:bg-blue-900 dark:border-gray-700 dark:text-gray-200 dark:placeholder-gray-400",
            size === "small" ? "h-9 pl-9 pr-4" : "h-14 pl-12 pr-4 shadow-sm",
          )}
        />

        <ComboboxOptions
          anchor="bottom start"
          className="z-10 mt-1 max-h-96 w-(--input-width) rounded-lg border border-gray-200 bg-white shadow-lg divide-y divide-gray-100 [--anchor-max-height:theme(height.96)] [--anchor-padding:theme(padding.4)] empty:hidden dark:border-gray-700 dark:bg-blue-900 dark:divide-gray-800"
        >
          {canShowResults
            ? filtered.map((item) => (
                <div key={item.type}>
                  <h2 className="px-4 py-2 text-sm font-semibold">
                    {item.label}
                  </h2>
                  {item.results.map((result) => (
                    <ComboboxOption
                      key={result.id}
                      value={result}
                      className="cursor-pointer px-4 py-2 data-focus:bg-brand-500 data-focus:text-inherit dark:data-focus:bg-brand-800"
                      as="div"
                    >
                      {(item.type === SearchResultType.Provider ||
                        item.type === SearchResultType.ProviderResource ||
                        item.type === SearchResultType.ProviderDatasource ||
                        item.type === SearchResultType.ProviderFunction) && (
                        <SearchProviderResult result={result} />
                      )}
                      {item.type === SearchResultType.Module && (
                        <SearchModuleResult result={result} />
                      )}
                      {item.type === SearchResultType.Other && (
                        <SearchOtherResult result={result} />
                      )}
                    </ComboboxOption>
                  ))}
                </div>
              ))
            : null}
          {!canShowResults && (
            <div className="flex h-12 items-center px-4 text-sm">
              {canShowLoadingInfo && (
                <Paragraph className="dark:text-white">Loading...</Paragraph>
              )}
              {canShowNoResultsInfo && <Paragraph>No results found</Paragraph>}
            </div>
          )}
        </ComboboxOptions>
      </div>
    </Combobox>
  );
}
