import clsx from "clsx";
import { BlockquoteHTMLAttributes, ReactNode } from "react";
import { isElement } from "react-is";

const WARNING_MARK = "[!WARNING]";
const IMPORTANT_MARK = "[!DANGER]";
const NOTE_MARK = "[!NOTE]";
const TIP_MARK = "[!TIP]";
const CAUTION_MARK = "[!CAUTION]";

function getAdmonitionClassName(prefix: string) {
  switch (prefix) {
    case NOTE_MARK:
      return "bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-100";
    case CAUTION_MARK:
      return "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-100";
    case WARNING_MARK:
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-100";
    case TIP_MARK:
      return "bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-100";
    case IMPORTANT_MARK:
      return "bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-100";
  }
}

function getAdmonitionTitle(prefix: string) {
  switch (prefix) {
    case NOTE_MARK:
      return "Note";
    case CAUTION_MARK:
      return "Caution";
    case WARNING_MARK:
      return "Warning";
    case TIP_MARK:
      return "Tip";
    case IMPORTANT_MARK:
      return "Important";
  }
}

function getAdmonitionMatch(children: ReactNode) {
  if (!Array.isArray(children)) {
    return null;
  }

  for (let i = 0; i < children.length; i++) {
    const child = children[i];

    if (!isElement(child)) {
      continue;
    }

    const type = child.props.children;

    if (typeof type !== "string") {
      continue;
    }

    switch (type) {
      case WARNING_MARK:
      case CAUTION_MARK:
      case NOTE_MARK:
      case TIP_MARK:
      case IMPORTANT_MARK:
        return {
          prefix: type,
          content: children.slice(i + 1),
        };
    }
  }

  return null;
}

export function MarkdownBlockquote({
  children,
}: BlockquoteHTMLAttributes<HTMLQuoteElement>) {
  const admonitionMatch = getAdmonitionMatch(children);

  if (admonitionMatch) {
    const { prefix, content } = admonitionMatch;
    const className = getAdmonitionClassName(prefix);
    const title = getAdmonitionTitle(prefix);

    return (
      <div
        className={clsx(
          "mt-5 px-3 py-2 first:mt-0 [&>details]:mt-2.5 [&>h6]:mt-4 [&>p]:mt-2.5 [li>&:first-child]:mt-0",
          className,
        )}
      >
        <strong className="font-semibold">{title}</strong>
        {content}
      </div>
    );
  }

  return (
    <blockquote className="mt-5 flex flex-col border-l-2 border-gray-500 pl-5">
      {children}
    </blockquote>
  );
}
