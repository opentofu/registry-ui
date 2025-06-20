import { NavLink } from "react-router";
import { TreeViewItem } from "@/components/TreeView";
import { ReactNode } from "react";

interface ModuleTabLinkProps {
  to: string;
  children: ReactNode;
  end?: boolean;
}

export function ModuleTabLink({ to, children, end }: ModuleTabLinkProps) {
  return (
    <TreeViewItem>
      <NavLink
        end={end}
        to={to}
        className={({ isActive }) =>
          `flex px-4 py-2 ${isActive ? "bg-brand-500 text-brand-600 text-inherit dark:bg-brand-800" : "text-inherit hover:bg-gray-100 dark:hover:bg-blue-900"}`
        }
      >
        {children}
      </NavLink>
    </TreeViewItem>
  );
}
