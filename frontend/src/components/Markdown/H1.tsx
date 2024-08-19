import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH1({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3 className="group scroll-mt-28 break-words font-bold text-4xl" id={id}>
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h3>
  );
}
