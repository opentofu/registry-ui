import { HTMLAttributes } from "react";

export function MarkdownUl({ children }: HTMLAttributes<HTMLUListElement>) {
  return (
    <ul className="my-4 ml-6 flex list-disc flex-col gap-2">{children}</ul>
  );
}
