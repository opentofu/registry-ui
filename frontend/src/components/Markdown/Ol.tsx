import { HTMLAttributes } from "react";

export function MarkdownOl({ children }: HTMLAttributes<HTMLUListElement>) {
  return (
    <ol className="my-4 ml-8 flex list-decimal flex-col gap-2">{children}</ol>
  );
}
