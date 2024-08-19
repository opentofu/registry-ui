import clsx from "clsx";
import { ReactNode } from "react";

interface TreeViewProps {
  children: ReactNode;
  className?: string;
}

export function TreeView({ children, className }: TreeViewProps) {
  return <ul className={clsx("flex flex-col", className)}>{children}</ul>;
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
        nested &&
          "border-l border-gray-300 content-none before:absolute before:-left-px before:-top-[2px] before:h-6 before:w-2 before:border-b before:border-l before:border-gray-300 last:border-transparent dark:border-gray-700 dark:before:border-gray-700 dark:last:border-transparent",
        className,
      )}
    >
      {children}
    </li>
  );
}
