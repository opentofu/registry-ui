import {
  Combobox,
  ComboboxInput,
  ComboboxOption,
  ComboboxOptions,
} from "@headlessui/react";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getSearchQuery } from "../../q";
import { search } from "../../icons/search";
import { spinner } from "../../icons/spinner";
import { Icon } from "../Icon";
import { SearchResult, SearchResultType } from "./types";
import { SearchModuleResult } from "./ModuleResult";
import { SearchProviderResult } from "./ProviderResult";
import { SearchOtherResult } from "./OtherResult";
import { definitions } from "@/api";
import clsx from "clsx";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { Paragraph } from "../Paragraph";

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
      if (event.key === "/" && event.target !== inputRef.current) {
        event.preventDefault();
        inputRef.current?.focus();
      }
    };

    document.addEventListener("keydown", handleSlash);

    return () => {
      document.removeEventListener("keydown", handleSlash);
    };
  }, []);

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
            "absolute z-10 text-gray-600",
            size === "small" ? "left-2 top-2 size-5" : "left-3 top-3 size-6",
          )}
        />
        <ComboboxInput
          ref={inputRef}
          value={query}
          onChange={(event) => onChange(event.target.value)}
          placeholder={placeholder}
          className={clsx(
            "relative block w-full appearance-none border border-transparent bg-gray-200 px-4 text-inherit placeholder:text-gray-500 focus:border-brand-700 focus:outline-none dark:bg-gray-800",
            size === "small" ? "h-9 pl-8" : "h-12 pl-10",
          )}
        />

        {isFetching && (
          <Icon
            path={spinner}
            className={clsx(
              "absolute animate-spin",
              size === "small"
                ? "right-2 top-2 size-5"
                : "right-3 top-3 size-6",
            )}
          />
        )}

        <ComboboxOptions
          anchor="bottom start"
          className="z-10 max-h-96 w-[var(--input-width)] divide-y divide-gray-300 bg-gray-200 [--anchor-max-height:theme(height.96)] [--anchor-padding:theme(padding.4)] dark:divide-gray-900 dark:bg-gray-800"
        >
          {filtered.map((item) => (
            <div key={item.type}>
              <h2 className="px-4 py-2 text-sm font-semibold">{item.label}</h2>
              {item.results.map((result) => (
                <ComboboxOption
                  key={result.id}
                  value={result}
                  className="cursor-pointer px-4 py-2 data-[focus]:bg-brand-500 data-[focus]:text-inherit dark:data-[focus]:bg-brand-800"
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
          ))}
          {filtered.length === 0 && !isFetching && query && (
            <Paragraph className="px-4 py-2 text-sm">
              No results found
            </Paragraph>
          )}
        </ComboboxOptions>
      </div>
    </Combobox>
  );
}
