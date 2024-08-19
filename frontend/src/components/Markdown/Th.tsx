import { HTMLAttributes } from "react";

export function MarkdownTh({ children }: HTMLAttributes<HTMLTableCellElement>) {
  return (
    <th className="border border-gray-200 p-2 dark:border-gray-800">
      {children}
    </th>
  );
}
