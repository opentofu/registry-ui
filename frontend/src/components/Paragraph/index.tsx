import clsx from "clsx";
import { ReactNode } from "react";

interface ParagraphProps {
  children: ReactNode;
  className?: string;
}

export function Paragraph({ children, className }: ParagraphProps) {
  return (
    <p className={clsx("text-gray-800 dark:text-gray-300", className)}>
      {children}
    </p>
  );
}
