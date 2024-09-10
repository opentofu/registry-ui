import clsx from "clsx";
import { ReactNode } from "react";

export enum AdmonitionType {
  NOTE = "note",
  CAUTION = "caution",
  WARNING = "warning",
  TIP = "tip",
  INFO = "info",
  IMPORTANT = "important",
}

function getAdmonitionClassName(type: AdmonitionType) {
  switch (type) {
    case AdmonitionType.NOTE:
      return "bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-100";
    case AdmonitionType.CAUTION:
      return "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-100";
    case AdmonitionType.WARNING:
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-100";
    case AdmonitionType.TIP:
      return "bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-100";
    case AdmonitionType.IMPORTANT:
    case AdmonitionType.INFO:
      return "bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-100";
  }
}

function getAdmonitionTitle(type: AdmonitionType) {
  switch (type) {
    case AdmonitionType.NOTE:
      return "Note";
    case AdmonitionType.CAUTION:
      return "Caution";
    case AdmonitionType.WARNING:
      return "Warning";
    case AdmonitionType.TIP:
      return "Tip";
    case AdmonitionType.IMPORTANT:
    case AdmonitionType.INFO:
      return "Important";
  }
}

interface AdmonitionProps {
  type?: AdmonitionType;
  children: ReactNode;
}

export function Admonition({ type, children }: AdmonitionProps) {
  const className = type ? getAdmonitionClassName(type) : null;

  return (
    <div
      className={clsx(
        "mt-5 px-3 py-2 first:mt-0 [&>details]:mt-2.5 [&>h6]:mt-4 [&>p]:mt-2.5 [li>&:first-child]:mt-0",
        className,
      )}
    >
      {type && (
        <strong className="font-semibold">{getAdmonitionTitle(type)}</strong>
      )}
      {children}
    </div>
  );
}
