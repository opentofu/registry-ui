import { HTMLAttributes } from "react";

export function MarkdownCode({ children }: HTMLAttributes<HTMLElement>) {
  return (
    <code className="bg-gray-100 px-1 py-0.5 font-mono text-sm text-gray-800 dark:bg-gray-800 dark:text-gray-200">
      {children}
    </code>
  );
}
