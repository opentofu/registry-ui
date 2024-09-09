import { HTMLAttributes } from "react";

export function MarkdownSummary({ children }: HTMLAttributes<HTMLElement>) {
  return (
    <summary className="cursor-default select-none font-semibold">
      {children}
    </summary>
  );
}
