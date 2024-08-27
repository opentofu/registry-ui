import { HTMLAttributes, ReactNode } from "react";
import { Paragraph } from "../Paragraph";
import clsx from "clsx";

const admonitionRegex = /^(?<prefix>!>|~>|->)\s+(?<text>.*)$/;

const getAdmonitionClassName = (prefix: string) => {
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
};

function getFirstLine(children: ReactNode) {
  if (typeof children === "string") {
    return children;
  }

  return Array.isArray(children) && typeof children[0] === "string"
    ? children[0]
    : null;
}

export function MarkdownP({ children }: HTMLAttributes<HTMLParagraphElement>) {
  const firstLine = getFirstLine(children);

  if (firstLine && admonitionRegex.test(firstLine)) {
    const match = firstLine.match(admonitionRegex);

    if (match?.groups) {
      const { prefix, text } = match.groups;
      const className = getAdmonitionClassName(prefix);
      const remainingLines = Array.isArray(children) ? children.slice(1) : null;

      return (
        <div
          role="alert"
          className={clsx("mt-5 px-3 py-2 [li>&:first-child]:mt-0", className)}
        >
          {text}
          {remainingLines}
        </div>
      );
    }
  }

  return (
    <Paragraph className="mt-5 leading-7 [li>&:first-child]:mt-0">
      {children}
    </Paragraph>
  );
}
