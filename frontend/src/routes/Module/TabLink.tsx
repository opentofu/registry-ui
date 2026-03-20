import { NavLink } from "react-router";
import { TreeViewItem } from "../../components/TreeView";
import { ReactNode } from "react";
import clsx from "clsx";
import { Icon } from "@/components/Icon";
import { lock } from "@/icons/lock";

interface ModuleTabLinkProps {
  to: string;
  children: ReactNode;
  end?: boolean;
  count?: number;
  disabled?: boolean;
}

export function ModuleTabLink({
  to,
  children,
  end,
  count,
  disabled,
}: ModuleTabLinkProps) {
  const sharedClasses = "flex px-3 py-2 items-center text-sm rounded-md transition-all duration-150";

  const component = disabled ? (
    <button
      disabled
      className={clsx(sharedClasses, "justify-between text-gray-400 dark:text-gray-500 cursor-not-allowed")}
    >
      {children}
      <Icon path={lock} className="size-em" />
    </button>
  ) : (
    <NavLink
      end={end}
      to={to}
      className={({ isActive }) =>
        clsx(
          sharedClasses,
          isActive
            ? "bg-brand-500/10 text-brand-700 dark:bg-brand-500/20 dark:text-brand-400 font-medium"
            : "text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white",
        )
      }
    >
      {children}
      {count !== undefined && ` (${count})`}
    </NavLink>
  );

  return <TreeViewItem>{component}</TreeViewItem>;
}
