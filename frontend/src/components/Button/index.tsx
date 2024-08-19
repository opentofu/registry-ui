import { MouseEventHandler, ReactNode } from "react";
import clsx from "clsx";
import { Link } from "react-router-dom";

interface RouterLinkProps {
  to: string;
}

interface NativeLinkProps {
  href: string;
}

interface NativeButtonProps {
  onClick: MouseEventHandler<HTMLButtonElement>;
}

type InteractionProps = RouterLinkProps | NativeLinkProps | NativeButtonProps;

interface BaseProps {
  variant: "primary" | "secondary";
  className?: string;
  children: ReactNode;
}

type ButtonProps = BaseProps & InteractionProps;

export function Button({ children, variant, className, ...rest }: ButtonProps) {
  const computedClassName = clsx(
    "border font-bold h-12 px-6 inline-flex items-center hover:no-underline transition-colors",
    variant === "primary" &&
      "bg-brand-500 text-gray-900 hover:bg-brand-600 border-brand-500 hover:border-brand-600 hover:text-gray-900",
    variant === "secondary" &&
      "border-gray-200 dark:border-gray-800 text-gray-900 dark:text-gray-50 bg-transparent hover:border-gray-900 dark:hover:border-gray-50 hover:text-gray-900 dark:hover:text-gray-50 aria-selected:border-gray-900 dark:aria-selected:border-gray-50",
    className,
  );

  if ("to" in rest && rest.to) {
    return (
      <Link {...rest} className={computedClassName}>
        {children}
      </Link>
    );
  }

  if ("href" in rest && rest.href) {
    return (
      <a {...rest} className={computedClassName}>
        {children}
      </a>
    );
  }

  return (
    <button {...rest} className={computedClassName}>
      {children}
    </button>
  );
}
