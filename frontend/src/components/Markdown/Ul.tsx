import { HTMLAttributes } from "react";

export function MarkdownUl({ children }: HTMLAttributes<HTMLUListElement>) {
  return (
    <ul className="ml-8 mt-5 flex list-disc flex-col gap-2 [li>&]:mt-2.5">
      {children}
    </ul>
  );
}
