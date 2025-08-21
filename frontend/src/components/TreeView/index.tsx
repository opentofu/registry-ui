import clsx from "clsx";
import { ReactNode } from "react";

interface TreeViewProps {
  children: ReactNode;
  className?: string;
}

export function TreeView({ children, className }: TreeViewProps) {
  return <ul className={clsx("flex flex-col space-y-0.5", className)}>{children}</ul>;
}

interface TreeViewItemProps {
  children: ReactNode;
  nested?: boolean;
  className?: string;
}

export function TreeViewItem({
  children,
  className,
  nested,
}: TreeViewItemProps) {
  return (
    <li
      className={clsx(
        "relative flex flex-col",
        nested && "ml-6",
        className,
      )}
    >
      {children}
    </li>
  );
}
