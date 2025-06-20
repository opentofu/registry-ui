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
    default:
      return "";
  }
}

function getAdmonitionMatch(children: ReactNode) {
  let text = "";

  if (typeof children === "string") {
    text = children;
  } else if (Array.isArray(children) && typeof children[0] === "string") {
    text = children[0];
  }

  if (!text) {
    return null;
  }

  if (
    text.startsWith(WARNING_MARK) ||
    text.startsWith(DANGER_MARK) ||
    text.startsWith(NOTE_MARK)
  ) {
    const content = text.slice(2);

    return {
      prefix: text.slice(0, 2),
      content: content.trim() ? content : null,
    };
  }

  return null;
}

interface MarkdownPProps extends HTMLAttributes<HTMLParagraphElement> {
  align?: string;
}

export function MarkdownP({ children, align }: MarkdownPProps) {
  const admonitionMatch = getAdmonitionMatch(children);

  if (admonitionMatch) {
    const { prefix, content } = admonitionMatch;
    const className = getAdmonitionClassName(prefix);
    const remainingContent = Array.isArray(children) ? children.slice(1) : null;

    return (
      <div
        role="alert"
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

  const alignClass = align === 'right' ? 'text-right' : align === 'center' ? 'text-center' : align === 'left' ? 'text-left' : '';

  return (
    <Paragraph className={clsx("mt-5 leading-7 first:mt-0 [li>&:first-child]:mt-0", alignClass)}>
      {children}
    </Paragraph>
  );
}
