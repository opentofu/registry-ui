import { Link, NavLink, UIMatch, useMatches } from "react-router";
import { Fragment } from "react";
import { chevron } from "../../icons/chevron";
import { Icon } from "../Icon";
import { home } from "../../icons/home";
import clsx from "clsx";
import { Crumb } from "../../crumbs";

interface BreadcrumbsProps {
  className?: string;
}

export function Breadcrumbs({ className }: BreadcrumbsProps) {
  const matches = useMatches() as Array<
    UIMatch<unknown, { crumb: (data: unknown) => Crumb }>
  >;

  console.log("Breadcrumbs matches:", matches);

  const crumbs = matches
    .filter((match) => Boolean(match.handle?.crumb))
    .map((match) => match.handle.crumb(match.data))
    .flat();

  return (
    <nav
      className={clsx("flex h-16 items-center space-x-2", className)}
      aria-label="Breadcrumbs"
    >
      <Link
        to="/"
        className="text-gray-700 dark:text-gray-300"
        aria-label="Home"
      >
        <Icon path={home} className="size-6" />
      </Link>
      {crumbs.map((crumb) => (
        <Fragment key={crumb.to}>
          <span>
            <Icon
              path={chevron}
              className="size-5 text-gray-400 dark:text-gray-600"
            />
          </span>
          <NavLink
            to={crumb.to}
            className="text-gray-700 hover:underline dark:text-gray-300"
            end
          >
            {crumb.label}
          </NavLink>
        </Fragment>
      ))}
    </nav>
  );
}

export function BreadcrumbsSkeleton() {
  return (
    <nav className="flex h-16 items-center">
      <div className="flex h-em w-80 animate-pulse bg-gray-500/25" />
    </nav>
  );
}
