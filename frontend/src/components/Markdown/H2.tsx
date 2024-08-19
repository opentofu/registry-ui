import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH2({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h4
      className="group mt-8 scroll-mt-28 break-words font-bold text-3xl"
      id={id}
    >
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h4>
  );
}
