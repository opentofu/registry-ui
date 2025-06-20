import { HeaderLogo } from "./Logo";
import { HeaderLink } from "./Link";
import { Icon } from "../Icon";
import { x } from "../../icons/x";
import { slack } from "../../icons/slack";
import { ThemeSwitcher } from "./ThemeSwitcher";
import { Search } from "../Search";
import { Link, useLocation } from "react-router";

export function Header() {
  // if we're on the home page, dont show the search bar
  const location = useLocation();
  const isHome = location.pathname === "/";

  const shouldShowSearch = !isHome;

  return (
    <header className="absolute top-12 left-0 right-0 z-50 flex h-20 items-center">
      <div className="flex items-center justify-between container mx-auto p-4 navbar">
        <div className="flex items-end gap-6">
          <h1>
            <Link
              to="/"
              aria-label="OpenTofu Registry"
              className="flex items-end transition-colors text-gray-900 hover:text-brand-500 dark:text-gray-50 dark:hover:text-brand-500"
            >
              <HeaderLogo />
              <span className="text-2xl tracking-wide">Registry</span>
            </Link>
          </h1>

          <nav className="flex items-end gap-6 pb-0.5">
            <HeaderLink to="/" label="Home" isActive={(id) => id === "home"} />
            <HeaderLink
              to="/providers"
              label="Providers"
              isActive={(id) => id.startsWith("provider")}
            />
            <HeaderLink
              to="/modules"
              label="Modules"
              isActive={(id) => id.startsWith("module")}
            />
            <HeaderLink
              to="/docs"
              label="Docs"
              isActive={(id) => id === "docs"}
            />
          </nav>
        </div>

        <nav className="flex items-center gap-6">
          {shouldShowSearch && <Search />}
          <a
            href="https://x.com/opentofuorg"
            aria-label="Follow us on X"
            target="_blank"
            className="transition-colors text-gray-900 hover:text-brand-500 dark:text-gray-50 dark:hover:text-brand-500"
          >
            <Icon path={x} className="size-6" />
          </a>
          <a
            href="https://opentofu.org/slack"
            aria-label="Join us on Slack"
            target="_blank"
            className="transition-colors text-gray-900 hover:text-brand-500 dark:text-gray-50 dark:hover:text-brand-500"
          >
            <Icon path={slack} className="size-6" />
          </a>
          <ThemeSwitcher />
        </nav>
      </div>
    </header>
  );
}
