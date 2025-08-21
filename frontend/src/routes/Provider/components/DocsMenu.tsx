import { Icon } from "@/components/Icon";
import { TreeView, TreeViewItem } from "@/components/TreeView";
import { chevron } from "@/icons/chevron";
import { useSuspenseQuery } from "@tanstack/react-query";
import clsx from "clsx";
import { useDeferredValue, useState, useTransition } from "react";
import { To, useHref, useLinkClickHandler } from "react-router";

import {
  NestedItem,
  filterSidebarItem,
  transformStructure,
} from "../docsSidebar";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderVersionDataQuery } from "../query";
import { Suspense } from "react";

type TabLinkProps = {
  to: To;
  label: string;
  active?: boolean;
};
function TabLink({ to, label, active }: TabLinkProps) {
  const href = useHref(to);
  const handleClick = useLinkClickHandler(to);
  const [isPending, startTransition] = useTransition();

  return (
    <a
      href={href}
      onClick={(event) => {
        if (!event.defaultPrevented) {
          startTransition(() => {
            handleClick(event);
          });
        }
      }}
      className={clsx(
        "flex rounded-md px-3 py-2 text-left text-sm break-all transition-all duration-150",
        active &&
        "bg-brand-300/40 text-brand-800 dark:bg-brand-300/40 dark:text-brand-200 font-medium",
        !active &&
        "text-gray-700 hover:bg-gray-200 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white",
        isPending && "opacity-50 transition-opacity delay-75",
      )}
    >
      {label}
    </a>
  );
}

type DocsTreeViewItemProps = {
  item: NestedItem;
  isOpenByDefault?: boolean;
  nested?: boolean;
  searchFilter?: string;
};
function DocsTreeViewItem({
  item,
  isOpenByDefault = false,
  nested = false,
  searchFilter = "",
}: DocsTreeViewItemProps) {
  const { lang } = useProviderParams();
  const [open, setOpen] = useState(isOpenByDefault);
  const filteredItems = item.items?.filter((subitem) =>
    filterSidebarItem(subitem, searchFilter),
  );

  let button;

  if (filteredItems) {
    button = (
      <button
        className="flex items-center gap-2 rounded-md px-3 py-2 text-left text-sm text-gray-700 transition-all duration-150 hover:bg-gray-200 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white"
        onClick={() => setOpen(!open)}
      >
        <Icon
          path={chevron}
          className={clsx(
            "size-3.5 shrink-0 transition-transform duration-200",
            open && "rotate-90",
          )}
        />
        <span className="font-medium">{item.label}</span>
      </button>
    );
  } else {
    button = (
      <TabLink
        to={{
          pathname: `docs/${item.type}/${item.name}`,
          search: lang ? `?lang=${lang}` : "",
        }}
        label={item.name}
        active={item.active}
      />
    );
  }

  return (
    <TreeViewItem nested={nested}>
      {button}
      {open && filteredItems && (
        <TreeView className="ml-4">
          {filteredItems.map((subitem) => (
            <DocsTreeViewItem
              key={subitem.name}
              item={subitem}
              isOpenByDefault
              nested
            />
          ))}
        </TreeView>
      )}
    </TreeViewItem>
  );
}

export function ProviderDocsMenu() {
  const { namespace, provider, version, doc, type, lang } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );
  const [searchFilter, setSearchFilter] = useState("");
  const deferredSearchFilter = useDeferredValue(searchFilter);

  const filterInput = (
    <Suspense>
      <input
        type="text"
        placeholder="Filter documentation..."
        className="focus:border-brand-500 focus:ring-brand-500/20 mb-4 h-10 w-full appearance-none rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-inherit transition-all duration-150 placeholder:text-gray-500 focus:ring-2 focus:outline-none dark:border-gray-700 dark:bg-gray-800/50 dark:placeholder:text-gray-400"
        value={deferredSearchFilter}
        onChange={(e) => setSearchFilter(e.target.value.toLocaleLowerCase())}
      />
    </Suspense>
  );
  const items = transformStructure(data.docs, type, doc);
  const filteredItems = items.filter((item) =>
    filterSidebarItem(item, searchFilter),
  );
  return (
    <div className="p-4">
      {filterInput}
      <TreeView>
        <TreeViewItem>
          <TabLink
            to={{
              pathname: `.`,
              search: lang ? `?lang=${lang}` : "",
            }}
            label="Overview"
            active={!type && !doc}
          />
        </TreeViewItem>
        {filteredItems.map((doc) => (
          <DocsTreeViewItem
            key={doc.name}
            item={doc}
            isOpenByDefault={doc.open}
            searchFilter={searchFilter}
          />
        ))}
      </TreeView>
    </div>
  );
}

export function ProviderDocsMenuSkeleton() {
  return (
    <div className="flex animate-pulse flex-col gap-5 p-4">
      <span className="h-em flex w-48 bg-gray-500/25" />
      <span className="h-em flex w-52 bg-gray-500/25" />
      <span className="h-em flex w-36 bg-gray-500/25" />
      <span className="h-em flex w-64 bg-gray-500/25" />
      <span className="h-em flex w-56 bg-gray-500/25" />
    </div>
  );
}
