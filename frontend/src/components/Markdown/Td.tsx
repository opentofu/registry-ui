import { HTMLAttributes } from "react";

export function MarkdownTd({ children }: HTMLAttributes<HTMLTableCellElement>) {
  return (
    <td className="border border-gray-200 p-2 dark:border-gray-800">
      {children}
    </td>
  );
}
