import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH3({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h5
      className="group mt-8 scroll-mt-28 break-words text-xl font-bold first:mt-0"
      id={id}
    >
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h5>
  );
}
