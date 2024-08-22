import {
  Combobox,
  ComboboxInput,
  ComboboxOption,
  ComboboxOptions,
} from "@headlessui/react";
import { useQuery } from "@tanstack/react-query";
import { useDeferredValue, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getSearchIndexQuery } from "../../q.ts";
import { search } from "../../icons/search";
import { spinner } from "../../icons/spinner.ts";
import { Icon } from "../Icon";
import useGroupedResults from "./hooks/useGroupedResults.tsx";

type SearchRefLink = {
  name: string;
  namespace: string;
  target_system?: string;
  version: string;
  id?: string;
};

enum SearchRefType {
  Module = "module",
  Provider = "provider",
  ProviderResource = "provider/resource",
  ProviderDatasource = "provider/datasource",
  ProviderFunction = "provider/function",
  Other = "other",
}

type SearchRef = {
  addr: string;
  type: SearchRefType;
  version: string;
  title: string;
  description: string;
  link: SearchRefLink;
  parent_id: string;
};

const parseRef = (ref: string): SearchRef => {
  return JSON.parse(ref);
};

const getTitle = (ref: SearchRef): string => {
  switch (ref.type) {
    case SearchRefType.Module:
      return `${ref.link.namespace}/${ref.link.name}`;
    case SearchRefType.Provider:
      return `${ref.link.namespace}/${ref.link.name}`;
    case SearchRefType.ProviderResource:
    case SearchRefType.ProviderDatasource:
    case SearchRefType.ProviderFunction:
      return `${ref.link.namespace}/${ref.link.name} - ${ref.link.id}`;
    default:
      return ref.title;
  }
};

const getLink = (ref: SearchRef): string => {
  switch (ref.type) {
    case SearchRefType.Module:
      return `/module/${ref.link.namespace}/${ref.link.name}/${ref.link.target_system}/${ref.link.version}`;
    case SearchRefType.Provider:
      return `/provider/${ref.link.namespace}/${ref.link.name}/${ref.link.version}`;
    case SearchRefType.ProviderResource:
      return `/provider/${ref.link.namespace}/${ref.link.name}/${ref.link.version}/docs/resources/${ref.link.id}`;
    case SearchRefType.ProviderDatasource:
      return `/provider/${ref.link.namespace}/${ref.link.name}/${ref.link.version}/docs/datasources/${ref.link.id}`;
    case SearchRefType.ProviderFunction:
      return `/provider/${ref.link.namespace}/${ref.link.name}/${ref.link.version}/docs/functions/${ref.link.id}`;
  }

  throw new Error(`Unknown ref type: ${ref.type}`);
};

export function Search() {
  const [query, setQuery] = useState("spacelift");
  const deferredQuery = useDeferredValue(query);
  const { data, isLoading } = useQuery(getSearchIndexQuery());
  const inputRef = useRef() as React.MutableRefObject<HTMLInputElement>;
  const navigate = useNavigate();

  const filtered = useGroupedResults(deferredQuery, data);

  const onChange = (value: string) => {
    setQuery(value);
  };

  useEffect(() => {
    const handleSlash = (event: KeyboardEvent) => {
      if (event.key === "/") {
        event.preventDefault();
        inputRef.current.focus();
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
      onChange={(v: { ref: string }) => {
        navigate(getLink(parseRef(v.ref)));
      }}
    >
      <div className="relative">
        <Icon
          path={search}
          className="absolute left-2 top-2 z-10 size-5 text-gray-600"
        />
        <ComboboxInput
          ref={inputRef}
          displayValue={(item: { namespace: string; name: string }) =>
            item?.name
          }
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
        className="z-10 max-h-96 w-96 bg-gray-200 [--anchor-max-height:theme(height.96)] empty:hidden dark:bg-gray-800"
      >
        {Array.from(filtered.entries()).map(([typeLabel, items]) => (
          <div key={typeLabel}>
            <h2 className="p-2 text-sm font-semibold">{typeLabel}</h2>
            {items.map((item) => (
              <ComboboxOption
                key={item.ref}
                value={item}
                className="cursor-pointer px-4 py-1 data-[focus]:bg-brand-500 data-[focus]:text-inherit dark:data-[focus]:bg-brand-800"
                as="div"
              >
                <div className="text-sm font-semibold">{getTitle(item)}</div>
                <div className="text-xs text-gray-500">{item.description}</div>
              </ComboboxOption>
            ))}
          </div>
        ))}
      </ComboboxOptions>
    </Combobox>
  );
}
