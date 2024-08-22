import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH2({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h4
      className="group mt-8 scroll-mt-5 break-words text-3xl font-bold first:mt-0"
      id={id}
    >
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h4>
  );
}
