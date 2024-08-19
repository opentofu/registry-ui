import { ReactNode } from "react";

interface CardItemProps {
  children: ReactNode;
}

export function CardItem({ children }: CardItemProps) {
  return (
    <article className="bg-gray-100 p-4 dark:bg-blue-900">{children}</article>
  );
}
