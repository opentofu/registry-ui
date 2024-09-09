import { HTMLAttributes } from "react";
import { HeadingLink } from "../HeadingLink";

export function MarkdownH4({
  children,
  id,
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h6
      className="group mt-8 scroll-mt-5 break-words text-base font-bold first:mt-0"
      id={id}
    >
      {children}
      {id && <HeadingLink id={id} label={children as string} />}
    </h6>
  );
}
