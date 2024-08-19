import clsx from "clsx";
import { ReactNode } from "react";

interface CardItemFooterProps {
  children: ReactNode;
}

export function CardItemFooter({ children }: CardItemFooterProps) {
  return (
    <footer>
      <dl className="flex gap-10">{children}</dl>
    </footer>
  );
}

interface CardItemFooterDetailProps {
  label: string;
  children: ReactNode;
  className?: string;
}

export function CardItemFooterDetail({
  label,
  children,
  className,
}: CardItemFooterDetailProps) {
  return (
    <div className={clsx("flex gap-2", className)}>
      <dt className="text-gray-700 dark:text-gray-300">{label}</dt>
      <dd>{children}</dd>
    </div>
  );
}
