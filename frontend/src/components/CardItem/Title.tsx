import { ReactNode } from "react";
import { Link } from "react-router-dom";

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
        {children}
      </Link>
    </h3>
  );
}
