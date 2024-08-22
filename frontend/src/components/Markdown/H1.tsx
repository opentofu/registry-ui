import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH1({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className="group mt-8 scroll-mt-5 break-words text-4xl font-bold first:mt-0"
      id={id}
    >
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h3>
  );
}
