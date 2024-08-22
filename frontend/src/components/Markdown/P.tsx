import { HTMLAttributes } from "react";
import { Paragraph } from "../Paragraph";

export function MarkdownP({ children }: HTMLAttributes<HTMLParagraphElement>) {
  return (
    <Paragraph className="mt-5 leading-7 [li>&:first-child]:mt-0">
      {children}
    </Paragraph>
  );
}
