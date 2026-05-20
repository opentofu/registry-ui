import clsx from "clsx";
import { ReactNode } from "react";

interface SidebarPanelProps {
  children: ReactNode;
  className?: string;
}

export function SidebarPanel({ children, className }: SidebarPanelProps) {
  return (
    <aside
      className={clsx(
        "custom-scrollbar sticky top-0 flex max-h-screen w-1/5 min-w-80 shrink-0 flex-col overflow-y-auto border-r border-gray-200 bg-gray-100 pt-2 dark:border-gray-800 dark:bg-blue-900",
        className,
      )}
    >
      {children}
    </aside>
  );
}
