import { Link, NavLink, UIMatch, useMatches } from "react-router";
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

  const crumbs = matches
    .filter((match) => Boolean(match.handle?.crumb))
    .map((match) => match.handle.crumb(match.data))
    .flat();

  return (
    <nav
      className={clsx("flex h-16 items-center", className)}
      aria-label="Breadcrumbs"
    >
      <ol className="flex items-center space-x-2">
        <li>
          <Link
            to="/"
            className="text-gray-700 dark:text-gray-300"
            aria-label="Home"
            aria-current={crumbs.length === 0 ? "page" : undefined}
          >
            <Icon path={home} className="size-6" />
          </Link>
        </li>

        {crumbs.map((crumb, index) => (
          <li key={crumb.to} className="flex items-center space-x-2">
            <span role="presentation">
              <Icon
                path={chevron}
                className="size-5 text-gray-400 dark:text-gray-600"
              />
            </span>
            {index === crumbs.length - 1 ? (
              // Use Link instead of NavLink for current page to preserve aria-current="page"
              // NavLink manages its own aria-current and overrides our explicit attribute
              <Link
                to={crumb.to}
                className="text-gray-700 hover:underline dark:text-gray-300"
                aria-current="page"
              >
                {crumb.label}
              </Link>
            ) : (
              <NavLink
                to={crumb.to}
                className="text-gray-700 hover:underline dark:text-gray-300"
                end
              >
                {crumb.label}
              </NavLink>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}

export function BreadcrumbsSkeleton() {
  return (
    <nav className="flex h-16 items-center">
      <div className="h-em flex w-80 animate-pulse bg-gray-500/25" />
    </nav>
  );
}
