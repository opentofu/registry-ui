import { ReactNode } from "react";

interface PageTitleProps {
  children: ReactNode;
  className?: string;
}

export function PageTitle({ children, className }: PageTitleProps) {
  return <h2 className={`${className} text-5xl font-bold`}>{children}</h2>;
}
