import { HeaderLogo } from "../Header/Logo";
import { Icon } from "../Icon";
import { x } from "../../icons/x";
import { slack } from "../../icons/slack";
import { home } from "../../icons/home";
import { document } from "../../icons/document";
import { ThemeSwitcher } from "../Header/ThemeSwitcher";
import { Search } from "../Search";
import { Link, useLocation, useMatches, UIMatch } from "react-router";
import { Breadcrumbs } from "../Breadcrumbs";
import { Crumb } from "../../crumbs";
import clsx from "clsx";

// File type icons for navigation tabs
const FileIcon = ({
	type,
}: {
	type: "home" | "providers" | "modules" | "docs";
}) => {
	const iconMap = {
		home: home,
		providers: target,
		modules: document,
		docs: document,
	};

	return <Icon path={iconMap[type]} className="size-4" />;
};

// Import target icon
import { target } from "../../icons/target";

interface TabLinkProps {
	to: string;
	icon: "home" | "providers" | "modules" | "docs";
	label: string;
	isActive: boolean;
}

function TabLink({ to, icon, label, isActive }: TabLinkProps) {
	return (
		<Link
			to={to}
			className={clsx(
				"flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors border-r border-gray-300 dark:border-gray-600",
				isActive
					? "bg-gray-100 dark:bg-blue-900 text-gray-900 dark:text-gray-100"
					: "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100/50 dark:hover:bg-blue-900/50",
			)}
		>
			<FileIcon type={icon} />
			<span>{label}</span>
			{isActive && <span className="text-xs opacity-60">●</span>}
		</Link>
	);
}

function BreadcrumbPath() {
	const matches = useMatches() as Array<
		UIMatch<unknown, { crumb: (data: unknown) => Crumb }>
	>;

	const crumbs = matches
		.filter((match) => Boolean(match.handle?.crumb))
		.map((match) => match.handle.crumb(match.data))
		.flat();

	const pathParts = ["~", "registry"];
	crumbs.forEach((crumb) => {
		pathParts.push(crumb.label.toLowerCase());
	});

	return (
		<div className="flex items-center gap-1 text-xs font-mono text-gray-600 dark:text-gray-400">
			<span>📁</span>
			<span>{pathParts.join("/")}</span>
		</div>
	);
}

export function IDEHeader() {
	const location = useLocation();
	const isHome = location.pathname === "/";
	const shouldShowSearch = !isHome;

	// Determine active tab based on current route
	const getActiveTab = () => {
		if (location.pathname === "/") return "home";
		if (location.pathname.startsWith("/providers")) return "providers";
		if (
			location.pathname.startsWith("/modules") ||
			location.pathname.startsWith("/module")
		)
			return "modules";
		if (location.pathname.startsWith("/docs")) return "docs";
		return "";
	};

	const activeTab = getActiveTab();

	return (
		<div className="bg-gray-200 dark:bg-blue-950 border border-gray-300 dark:border-gray-700 rounded-t">
			<div className="flex items-center justify-between px-4 py-2 bg-gray-100 dark:bg-blue-900 border-b border-gray-300 dark:border-gray-600">
				<div className="flex items-center gap-3">
					<div className="flex items-center gap-2">
						{/* TODO: Rework this, its messy righ now */}
						<HeaderLogo />
						<h1 className="text-sm font-medium text-gray-900 dark:text-gray-100">
							OpenTofu Registry
						</h1>
					</div>
				</div>

				<div className="flex items-center gap-4">
					{shouldShowSearch && <Search />}
					<a
						href="https://x.com/opentofuorg"
						aria-label="Follow us on X"
						target="_blank"
						className="transition-colors text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100"
					>
						<Icon path={x} className="size-4" />
					</a>
					<a
						href="https://opentofu.org/slack"
						aria-label="Join us on Slack"
						target="_blank"
						className="transition-colors text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100"
					>
						<Icon path={slack} className="size-4" />
					</a>
					<ThemeSwitcher />
				</div>
			</div>

			{/* Tab bar with breadcrumbs */}
			{/* TODO: Replace icons with proper ones */}
			<div className="flex items-center justify-between">
				<div className="flex items-center">
					<TabLink
						to="/"
						icon="home"
						label="Home"
						isActive={activeTab === "home"}
					/>
					<TabLink
						to="/providers"
						icon="providers"
						label="Providers"
						isActive={activeTab === "providers"}
					/>
					<TabLink
						to="/modules"
						icon="modules"
						label="Modules"
						isActive={activeTab === "modules"}
					/>
					<TabLink
						to="/docs"
						icon="docs"
						label="Docs"
						isActive={activeTab === "docs"}
					/>

					{/* Small spacer div */}
					<div className="w-4" />

					{/* File path breadcrumbs */}
					<BreadcrumbPath />
				</div>
			</div>
		</div>
	);
}
