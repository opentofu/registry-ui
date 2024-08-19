import { ReactNode } from "react";

interface PageTitleProps {
  children: ReactNode;
}

export function PageTitle({ children }: PageTitleProps) {
  return <h2 className="text-5xl font-bold">{children}</h2>;
}
