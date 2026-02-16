import { Icon } from "@/components/Icon";
import { TreeView, TreeViewItem } from "@/components/TreeView";
import { chevron } from "@/icons/chevron";

import clsx from "clsx";
import { useState } from "react";
import { NavLink, To } from "react-router";
import sidebar from "../../../../docs/sidebar.json";
import { SidebarItem } from "../types";

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
          "flex rounded-md px-3 py-2 text-left text-sm break-all transition-all duration-150",
          isActive &&
          "bg-brand-300/40 text-brand-800 dark:bg-brand-300/40 dark:text-brand-200 font-medium",
          !isActive &&
          "text-gray-700 hover:bg-gray-200 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white",
        )
      }
    >
      {label}
    </NavLink>
  );
}

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
        <span className="font-medium">{item.title}</span>
      </button>
    );
  } else {
    button = <TabLink to={`/docs/${item.slug}`} label={item.title} />;
  }

  return (
    <TreeViewItem nested={nested}>
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
    <div className="p-4">
      <TreeView>
        {sidebar.map((item) => (
          <DocsTreeViewItem key={item.title} item={item} isOpenByDefault />
        ))}
      </TreeView>
    </div>
  );
}
