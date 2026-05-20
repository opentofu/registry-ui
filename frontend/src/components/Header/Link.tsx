import clsx from "clsx";
import { Link, useMatches } from "react-router";

interface LinkProps {
  label: string;
  to: string;
  isActive: (routeId: string) => boolean;
}

export function HeaderLink({ label, to, isActive }: LinkProps) {
  const matches = useMatches();

  const isActiveMatch = !!matches.find((match) => isActive(match.id));

  return (
    <Link
      to={to}
      className={clsx(
        "font-semibold transition-colors text-gray-900 hover:text-brand-500 dark:text-gray-50 dark:hover:text-brand-500",
        isActiveMatch && "underline underline-offset-4",
      )}
    >
      {label}
    </Link>
  );
}
