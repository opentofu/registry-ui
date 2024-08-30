import { HeaderLogo } from "./Logo";
import { HeaderLink } from "./Link";
import { Icon } from "../Icon";
import { x } from "../../icons/x";
import { slack } from "../../icons/slack";
import { ThemeSwitcher } from "./ThemeSwitcher";
import { Search } from "../Search";
import { useLocation } from "react-router-dom";

export function Header() {
  // if we're on the home page, dont show the search bar
  const location = useLocation();
  const isHome = location.pathname === "/";

  const shouldShowSearch = !isHome;

  return (
    <header className="flex h-20 items-center border-b border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-blue-950">
      <div className="mx-auto flex w-full max-w-screen-3xl items-end px-5">
        <h1 className="flex items-end">
          <a href="/" aria-label="Home" target="_blank">
            <HeaderLogo />
          </a>
          <span className="text-2xl tracking-wide">Registry</span>
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
