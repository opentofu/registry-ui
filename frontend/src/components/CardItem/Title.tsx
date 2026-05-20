import { ReactNode } from "react";
import { Link } from "react-router";

interface CardItemTitleProps {
  children: ReactNode;
  linkProps: {
    to: string;
  };
  className?: string;
}

export function CardItemTitle({ children, linkProps, className }: CardItemTitleProps) {
  return (
    <h3 className={className}>
      <Link {...linkProps} className="text-xl font-semibold">
        <span aria-hidden className="absolute inset-0"></span>
        {children}
      </Link>
    </h3>
  );
}
