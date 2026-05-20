import { Icon } from "@/components/Icon";
import { home } from "@/icons/home";
import { chevron } from "@/icons/chevron";
import { Link } from "react-router";

interface SearchBreadcrumbsProps {
  onHomeClick: () => void;
}

// Simple hardcoded breadcrumb for search UI - purely presentational to match the rest of the UI
// Unlike the main Breadcrumbs component, this is not powered by routing logic for simplicity
export function SearchBreadcrumbs({ onHomeClick }: SearchBreadcrumbsProps) {
  return (
    <div className="flex h-10 items-center rounded-t border border-b-0 border-gray-300 bg-gray-200 px-3 dark:border-gray-700 dark:bg-blue-950">
      <nav
        className="flex h-10 items-center"
        aria-label="Breadcrumbs"
      >
        <ol className="flex items-center space-x-2">
          <li>
            <Link
              to="/"
              className="text-gray-700 dark:text-gray-300"
              aria-label="Home"
              onClick={(e) => {
                e.preventDefault();
                onHomeClick();
              }}
            >
              <Icon path={home} className="size-5" />
            </Link>
          </li>
          <li className="flex items-center space-x-2">
            <span role="presentation">
              <Icon
                path={chevron}
                className="size-4 text-gray-400 dark:text-gray-600"
              />
            </span>
            <span className="text-sm text-gray-700 dark:text-gray-300" aria-current="page">
              Search
            </span>
          </li>
        </ol>
      </nav>
    </div>
  );
}