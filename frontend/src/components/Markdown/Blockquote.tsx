import { BlockquoteHTMLAttributes } from "react";

export function MarkdownBlockquote({
  children,
}: BlockquoteHTMLAttributes<HTMLQuoteElement>) {
  return (
    <blockquote className="mt-5 flex flex-col border-l-2 border-gray-500 pl-5">
      {children}
    </blockquote>
  );
}
