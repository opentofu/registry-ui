import { ReactNode } from "react";
import { Link } from "react-router";

interface CardItemTitleProps {
  children: ReactNode;
  linkProps: {
    to: string;
  };
}

export function CardItemTitle({ children, linkProps }: CardItemTitleProps) {
  return (
    <h3>
      <Link {...linkProps} className="text-xl font-semibold">
        <span aria-hidden className="absolute inset-0"></span>
        {children}
      </Link>
    </h3>
  );
}
