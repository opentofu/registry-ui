import { HTMLAttributes } from "react";

export function MarkdownUl({ children }: HTMLAttributes<HTMLUListElement>) {
  return (
    <ul className="mt-5 ml-8 flex list-disc flex-col gap-2 [li>&]:mt-2.5">
      {children}
    </ul>
  );
}
