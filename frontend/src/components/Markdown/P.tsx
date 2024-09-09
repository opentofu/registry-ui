import { HTMLAttributes, ReactNode } from "react";
import { Paragraph } from "../Paragraph";
import clsx from "clsx";

const WARNING_MARK = "~>";
const DANGER_MARK = "!>";
const NOTE_MARK = "->";

function getAdmonitionClassName(prefix: string) {
  switch (prefix) {
    case NOTE_MARK:
      return "bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-100";
    case DANGER_MARK:
      return "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-100";
    case WARNING_MARK:
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-100";
  }
}

function getAdmonitionMatch(children: ReactNode) {
  let content = "";

  if (typeof children === "string") {
    content = children;
  } else if (Array.isArray(children) && typeof children[0] === "string") {
    content = children[0];
  }

  if (!content) {
    return null;
  }

  if (
    content.startsWith(WARNING_MARK) ||
    content.startsWith(DANGER_MARK) ||
    content.startsWith(NOTE_MARK)
  ) {
    return {
      prefix: content.slice(0, 2),
      content: content.slice(2).trim(),
    };
  }

  return null;
}

export function MarkdownP({ children }: HTMLAttributes<HTMLParagraphElement>) {
  const admonitionMatch = getAdmonitionMatch(children);

  if (admonitionMatch) {
    const { prefix, content } = admonitionMatch;
    const className = getAdmonitionClassName(prefix);
    const remainingContent = Array.isArray(children) ? children.slice(1) : null;

    return (
      <div
        className={clsx(
          "mt-5 px-3 py-2 first:mt-0 [li>&:first-child]:mt-0",
          className,
        )}
      >
        {content}
        {remainingContent}
      </div>
    );
  }

  return (
    <Paragraph className="mt-5 leading-7 first:mt-0 [li>&:first-child]:mt-0">
      {children}
    </Paragraph>
  );
}
