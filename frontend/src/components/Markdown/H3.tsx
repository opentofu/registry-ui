import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH3({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h5
      className="group mt-8 scroll-mt-5 text-xl font-bold break-words first:mt-0"
      id={id}
    >
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h5>
  );
}
