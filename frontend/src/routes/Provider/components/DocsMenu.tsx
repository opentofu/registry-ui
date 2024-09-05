import { Icon } from "@/components/Icon";
import { TreeView, TreeViewItem } from "@/components/TreeView";
import { chevron } from "@/icons/chevron";
import { useSuspenseQuery } from "@tanstack/react-query";
import clsx from "clsx";
import { useId, useState, useTransition } from "react";
import { To, useHref, useLinkClickHandler } from "react-router-dom";

import { NestedItem, transformStructure } from "../docsSidebar";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderVersionDataQuery } from "../query";

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
        "flex break-all px-4 py-2 text-left",
        active && "bg-brand-500 text-brand-600 text-inherit dark:bg-brand-800",
        !active && "text-inherit hover:bg-gray-100 dark:hover:bg-blue-900",
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
};
function DocsTreeViewItem({
  item,
  isOpenByDefault = false,
  nested = false,
}: DocsTreeViewItemProps) {
  const { lang } = useProviderParams();
  const [open, setOpen] = useState(isOpenByDefault);
  let button;
  const listId = useId();
  const buttonId = useId();

  if (item.items) {
    button = (
      <button
        className="flex gap-2 px-4 py-2 text-left text-inherit hover:bg-gray-100 dark:hover:bg-blue-900"
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-controls={open ? listId : undefined}
        id={buttonId}
      >
        <Icon
          path={chevron}
          className={clsx("mt-1 size-4 shrink-0", open && "rotate-90")}
        />
        {item.label}
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
    <TreeViewItem nested={nested} className={nested ? "ml-2" : ""}>
      {button}
      {open && item.items && (
        <TreeView className="ml-4" id={listId} aria-labelledby={buttonId}>
          {item.items.map((subitem) => (
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

  const items = transformStructure(data.docs, type, doc);

  return (
    <TreeView className="mr-4 mt-4">
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
      {items.map((doc) => (
        <DocsTreeViewItem
          key={doc.name}
          item={doc}
          isOpenByDefault={doc.open}
        />
      ))}
    </TreeView>
  );
}

export function ProviderDocsMenuSkeleton() {
  return (
    <div className="mr-4 mt-4 flex animate-pulse flex-col gap-5">
      <span className="flex h-em w-48 bg-gray-500/25" />
      <span className="flex h-em w-52 bg-gray-500/25" />
      <span className="flex h-em w-36 bg-gray-500/25" />
      <span className="flex h-em w-64 bg-gray-500/25" />
      <span className="flex h-em w-56 bg-gray-500/25" />
    </div>
  );
}
