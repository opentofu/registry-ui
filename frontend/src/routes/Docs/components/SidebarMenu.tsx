import { Icon } from "@/components/Icon";
import { TreeView, TreeViewItem } from "@/components/TreeView";
import { chevron } from "@/icons/chevron";

import clsx from "clsx";
import { useState } from "react";
import { NavLink, To } from "react-router-dom";
import sidebar from "../../../../docs/sidebar.json";

type TabLinkProps = {
  to: To;
  label: string;
};

function TabLink({ to, label }: TabLinkProps) {
  return (
    <NavLink
      end
      to={to}
      className={({ isActive }) =>
        clsx(
          "flex break-all px-4 py-2 text-left",
          isActive &&
            "bg-brand-500 text-brand-600 text-inherit dark:bg-brand-800",
          !isActive && "text-inherit hover:bg-gray-100 dark:hover:bg-blue-900",
        )
      }
    >
      {label}
    </NavLink>
  );
}

type SidebarItem =
  | {
      title: string;
      items: SidebarItem[];
    }
  | {
      title: string;
      slug: string;
      path: string;
      items?: never;
    };

type DocsTreeViewItemProps = {
  item: SidebarItem;
  isOpenByDefault?: boolean;
  nested?: boolean;
};

function DocsTreeViewItem({
  item,
  isOpenByDefault = false,
  nested = false,
}: DocsTreeViewItemProps) {
  const [open, setOpen] = useState(isOpenByDefault);
  let button;

  if (item.items) {
    button = (
      <button
        className="flex gap-2 px-4 py-2 text-left text-inherit hover:bg-gray-100 dark:hover:bg-blue-900"
        onClick={() => setOpen(!open)}
      >
        <Icon
          path={chevron}
          className={clsx("mt-1 size-4 shrink-0", open && "rotate-90")}
        />
        {item.title}
      </button>
    );
  } else {
    button = (
      <TabLink
        to={{
          pathname: item.slug,
        }}
        label={item.title}
      />
    );
  }

  return (
    <TreeViewItem nested={nested} className={nested ? "ml-2" : ""}>
      {button}
      {open && item.items && (
        <TreeView className="ml-4">
          {item.items.map((subitem) => (
            <DocsTreeViewItem
              key={subitem.title}
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

export function DocsSidebarMenu() {
  return (
    <TreeView className="mr-4 mt-4">
      {sidebar.map((item) => (
        <DocsTreeViewItem key={item.title} item={item} isOpenByDefault />
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
