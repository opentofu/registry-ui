import { AnchorHTMLAttributes } from "react";

function isExternalLink(href: string) {
  return href.startsWith("http");
}

export function MarkdownA({
  children,
  href,
  ...rest
}: AnchorHTMLAttributes<HTMLAnchorElement>) {
  return (
    <a
      href={href}
      className="text-gray-900 underline underline-offset-2 dark:text-gray-200"
      rel={href && isExternalLink(href) ? "noreferrer" : undefined}
      {...rest}
    >
      {children}
    </a>
  );
}
