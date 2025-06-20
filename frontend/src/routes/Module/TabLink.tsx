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
  const sharedClasses = "flex px-4 py-2 items-center";

  const component = disabled ? (
    <button
      disabled
      className={clsx(sharedClasses, "justify-between text-gray-500")}
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
            ? "bg-brand-500 text-brand-600 text-inherit dark:bg-brand-800"
            : "text-inherit hover:bg-gray-100 dark:hover:bg-blue-900",
        )
      }
    >
      {children}
      {count !== undefined && ` (${count})`}
    </NavLink>
  );

  return <TreeViewItem>{component}</TreeViewItem>;
}
