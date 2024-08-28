import { HTMLAttributes } from "react";

export function MarkdownOl({ children }: HTMLAttributes<HTMLUListElement>) {
  return (
    <ol className="ml-8 mt-5 flex list-decimal flex-col gap-2 [li>&]:mt-2.5">
      {children}
    </ol>
  );
}
