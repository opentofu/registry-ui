import {
  Combobox,
  ComboboxInput,
  ComboboxOption,
  ComboboxOptions,
} from "@headlessui/react";
import { useQuery } from "@tanstack/react-query";
import { useDeferredValue, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getSearchQuery } from "../../q";
import { search } from "../../icons/search";
import { spinner } from "../../icons/spinner";
import { Icon } from "../Icon";
import { ApiSearchResult, SearchResult, SearchResultType } from "./types";
import { SearchModuleResult } from "./ModuleResult";
import { SearchProviderResult } from "./ProviderResult";
import { SearchOtherResult } from "./OtherResult";

const getTypeOrder = (type: SearchResultType) => {
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
};

const getTypeLabel = (type: SearchResultType) => {
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
};

type Results = Array<{
  label: string;
  type: SearchResultType;
  results: SearchResult[];
}>;

export function Search() {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data, isLoading } = useQuery(getSearchQuery(deferredQuery));
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

      const result = data[i] as ApiSearchResult;
      const order = getTypeOrder(result.type);

      const group = (results[order] = results[order] || {
        type: result.type,
        label: getTypeLabel(result.type),
        results: [],
      });

      let link = "";

      if (result.type === SearchResultType.Module) {
        link = `/module/${result.link_variables.namespace}/${result.link_variables.name}/${result.link_variables.target_system}/${result.version}`;
      } else if (result.type === SearchResultType.Provider) {
        link = `/provider/${result.link_variables.namespace}/${result.link_variables.name}/${result.version}`;
      }

      group.results.push({
        id: result.id,
        title: result.title,
        addr: result.addr,
        description: result.description,
        link,
        type: result.type,
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
      onClose={() => setQuery("")}
      onChange={(v: SearchResult) => {
        navigate(v.link);
      }}
    >
      <div className="relative">
        <Icon
          path={search}
          className="absolute left-2 top-2 z-10 size-5 text-gray-600"
        />
        <ComboboxInput
          ref={inputRef}
          displayValue={(result: SearchResult) => result.title}
          onChange={(event) => onChange(event.target.value)}
          placeholder="Search resources (Press / to focus)"
          className="relative block h-9 w-96 appearance-none border border-transparent bg-gray-200 px-4 pl-8 text-inherit placeholder:text-gray-500 focus:border-brand-700 focus:outline-none dark:bg-gray-800"
        />

        {isLoading && (
          <Icon
            path={spinner}
            className="absolute right-2 top-2 size-5 animate-spin"
          />
        )}
      </div>
      <ComboboxOptions
        anchor="bottom start"
        className="z-10 max-h-96 w-96 divide-y divide-gray-300 bg-gray-200 [--anchor-max-height:theme(height.96)] empty:hidden dark:divide-gray-900 dark:bg-gray-800"
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
      </ComboboxOptions>
    </Combobox>
  );
}
