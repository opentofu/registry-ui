import { ReactNode } from "react";

interface CardItemProps {
  children: ReactNode;
}

export function CardItem({ children }: CardItemProps) {
  return (
    <article className="relative bg-white p-4 rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow dark:bg-blue-900 dark:border-gray-700 h-full w-full flex flex-col">
      {children}
    </article>
  );
}
