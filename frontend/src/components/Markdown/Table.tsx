import { HTMLAttributes } from "react";

export function MarkdownTable({ children }: HTMLAttributes<HTMLTableElement>) {
  return (
    <div className="overflow-x-auto">
      <table className="mt-4 table-fixed border-collapse border border-gray-200 dark:border-gray-800">
        {children}
      </table>
    </div>
  );
}
