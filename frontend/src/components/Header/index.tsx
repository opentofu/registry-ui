import { Link, useLocation } from "react-router";
import { useEffect, useRef } from "react";

import { HeaderLink } from "./Link";
import { HeaderLogo } from "./Logo";
import { Icon } from "../Icon";
import { Search } from "../Search";
import { ThemeSwitcher } from "./ThemeSwitcher";
import { slack } from "../../icons/slack";
import { x } from "../../icons/x";

export function Header() {
  // if we're on the home page, dont show the search bar
  const location = useLocation();
  const isHome = location.pathname === "/";

  // Scroll to an anchor after a page is loaded
  const lastHash = useRef('');
  useEffect(() => {
    if (location.hash) {
      lastHash.current = location.hash.slice(1);  // Get the hash from the URL
    }

    if (lastHash.current && document.getElementById(lastHash.current)) {
      setTimeout(() => {
        document
          .getElementById(lastHash.current)
          ?.scrollIntoView({ behavior: "smooth", block: "start" });
        lastHash.current = "";
      }, 100);
    }
  }, [location]);

  const shouldShowSearch = !isHome;

  return (
    <header className="flex h-20 items-center border-b border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-blue-950">
      <div className="mx-auto flex w-full max-w-(--breakpoint-3xl) items-end px-5">
        <h1>
          <Link
            to="/"
            aria-label="OpenTofu Registry"
            className="hover:text-brand-500 flex items-end"
          >
            <HeaderLogo />
            <span className="text-2xl tracking-wide">Registry</span>
          </Link>
        </h1>

        <nav className="ml-6 flex h-9 items-center gap-6">
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

        <nav className="ml-auto flex h-9 items-center gap-6">
          {shouldShowSearch && <Search />}
          <a
            href="https://x.com/opentofuorg"
            aria-label="Follow us on X"
            target="_blank"
          >
            <Icon path={x} className="size-6" />
          </a>
          <a
            href="https://opentofu.org/slack"
            aria-label="Join us on Slack"
            target="_blank"
          >
            <Icon path={slack} className="size-6" />
          </a>
          <ThemeSwitcher />
        </nav>
      </div>
    </header>
  );
}
