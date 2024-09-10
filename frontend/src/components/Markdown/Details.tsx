import { DetailsHTMLAttributes } from "react";

export function MarkdownDetails({
  children,
}: DetailsHTMLAttributes<HTMLDetailsElement>) {
  return <details className="mt-5">{children}</details>;
}
