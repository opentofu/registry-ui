import { HTMLAttributes } from "react";
import { Paragraph } from "../Paragraph";

export function MarkdownP({ children }: HTMLAttributes<HTMLParagraphElement>) {
  return <Paragraph className="mt-5">{children}</Paragraph>;
}
