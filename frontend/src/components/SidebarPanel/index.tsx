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
        "sticky top-20 flex max-h-[calc(100vh-theme(height.20)-1px)] w-1/5 min-w-80 shrink-0 flex-col overflow-y-auto [scrollbar-width:thin]",
        className,
      )}
    >
      {children}
    </aside>
  );
}
