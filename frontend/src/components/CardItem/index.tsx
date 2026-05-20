import { ReactNode } from "react";

interface CardItemProps {
  children: ReactNode;
}

export function CardItem({ children }: CardItemProps) {
  return (
    <article className="relative flex h-full w-full flex-col rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md dark:border-gray-700 dark:bg-blue-900">
      {children}
    </article>
  );
}
