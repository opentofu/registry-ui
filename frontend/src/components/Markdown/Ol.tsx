import { HTMLAttributes } from "react";

export function MarkdownOl({ children }: HTMLAttributes<HTMLUListElement>) {
  return (
    <ol className="mt-5 ml-8 flex list-decimal flex-col gap-2 [li>&]:mt-2.5">
      {children}
    </ol>
  );
}
