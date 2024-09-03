import { HTMLAttributes, ReactNode } from "react";
import { Paragraph } from "../Paragraph";
import clsx from "clsx";

const admonitionRegex = /^(?<prefix>!>|~>|->)\s+(?<content>.*)$/;

function getAdmonitionClassName(prefix: string) {
  switch (prefix) {
    case "->":
      return "bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-100";
    case "!>":
      return "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-100";
    case "~>":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-100";
    default:
      return "";
  }
}

function getAdmonitionMatch(children: ReactNode) {
  let content = "";

  if (typeof children === "string") {
    content = children;
  } else if (Array.isArray(children) && typeof children[0] === "string") {
    content = children[0];
  }

  const match = content.match(admonitionRegex);

  if (!match || !match.groups) {
    return null;
  }

  return {
    prefix: match.groups.prefix,
    content: match.groups.content,
  };
}

export function MarkdownP({ children }: HTMLAttributes<HTMLParagraphElement>) {
  const admonitionMatch = getAdmonitionMatch(children);

  if (admonitionMatch) {
    const { prefix, content } = admonitionMatch;
    const className = getAdmonitionClassName(prefix);
    const remainingContent = Array.isArray(children) ? children.slice(1) : null;

    return (
      <div
        role="alert"
        className={clsx("mt-5 px-3 py-2 [li>&:first-child]:mt-0", className)}
      >
        {content}
        {remainingContent}
      </div>
    );
  }

  return (
    <Paragraph className="mt-5 leading-7 [li>&:first-child]:mt-0">
      {children}
    </Paragraph>
  );
}
