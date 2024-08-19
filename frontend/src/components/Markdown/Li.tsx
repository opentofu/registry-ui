import { HTMLAttributes } from "react";

export function MarkdownLi({ children }: HTMLAttributes<HTMLLIElement>) {
  return <li className="text-gray-700 dark:text-gray-300">{children}</li>;
}
