import clsx from "clsx";
import { ReactNode } from "react";

interface CardItemFooterProps {
  children: ReactNode;
}

export function CardItemFooter({ children }: CardItemFooterProps) {
  return (
    <footer>
      <dl className="flex">{children}</dl>
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
    <div className={clsx("flex w-52 gap-2", className)}>
      <dt className="text-gray-700 dark:text-gray-300">{label}</dt>
      <dd>{children}</dd>
    </div>
  );
}

export function CardItemFooterDetailSkeleton() {
  return (
    <div className="flex w-52 gap-2">
      <span className="flex h-em w-28 animate-pulse bg-gray-500/25" />
      <span className="flex h-em w-10 animate-pulse bg-gray-500/25" />
    </div>
  );
}
