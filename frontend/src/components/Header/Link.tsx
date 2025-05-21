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
        "font-semibold",
        isActiveMatch && "underline underline-offset-2",
      )}
    >
      {label}
    </Link>
  );
}
