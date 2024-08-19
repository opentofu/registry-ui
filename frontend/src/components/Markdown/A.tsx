import { HTMLAttributes } from "react";

export function MarkdownA({
  children,
  ...rest
}: HTMLAttributes<HTMLAnchorElement>) {
  return (
    <a
      className="text-gray-900 underline underline-offset-2 dark:text-gray-200"
      target="_blank"
      rel="noopener noreferrer"
      {...rest}
    >
      {children}
    </a>
  );
}
