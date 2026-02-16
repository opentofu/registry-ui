import { ReactNode } from "react";
import { Footer } from "../Footer";
import { Header } from "../Header";
import { IDEHeader } from "../IDEHeader";
import { Breadcrumbs } from "../Breadcrumbs";
import PatternBg from "../PatternBg";

interface UnifiedLayoutProps {
	children: ReactNode;
	sidebar?: ReactNode;
	afterSidebar?: ReactNode;
	showWindowFrame?: boolean;
	showFooter?: boolean;
	showKeyboardShortcuts?: boolean;
	useIDEHeader?: boolean;
}

export function UnifiedLayout({
	children,
	sidebar,
	afterSidebar,
	showWindowFrame = true,
	showFooter = true,
	showKeyboardShortcuts = false,
	useIDEHeader = false,
}: UnifiedLayoutProps) {
	if (!showWindowFrame) {
		return (
			<>
				<PatternBg />
				{useIDEHeader ? <IDEHeader /> : <Header />}
				{children}
				{showFooter && <Footer />}
			</>
		);
	}

	return (
		<>
			<PatternBg />
			<div className="fixed inset-0 -z-10 bg-white/50 dark:bg-blue-950/50" />
			
			{!useIDEHeader && <Header />}

			<div className={`mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 ${useIDEHeader ? 'pt-12' : 'pt-24'}`}>
				{useIDEHeader ? (
					<IDEHeader />
				) : (
					<div className="h-10 bg-gray-200 dark:bg-blue-950 border border-gray-300 dark:border-gray-700 border-b flex items-center px-3 rounded-t">
						<Breadcrumbs className="h-10 flex-1" />
					</div>
				)}

				<div
					className={`flex flex-1 ${sidebar || afterSidebar ? "divide-x divide-gray-200 dark:divide-gray-800" : ""} border border-gray-300 dark:border-gray-700 border-t-0 ${!showKeyboardShortcuts ? "rounded-b" : ""}`}
				>
					{sidebar}
					<main
						className={`min-w-0 flex-1 bg-gray-100 dark:bg-blue-900 ${!sidebar && !afterSidebar ? "pb-5" : ""}`}
					>
						<div className="mt-8">{children}</div>
					</main>
					{afterSidebar}
				</div>

				{showKeyboardShortcuts && (sidebar || afterSidebar) && (
					<div className="flex h-8 items-center rounded-b border border-t-0 border-gray-300 bg-gray-200 px-3 dark:border-gray-700 dark:bg-blue-950">
						<div className="w-full text-center">
							<p className="text-xs text-gray-600 dark:text-gray-400">
								Use{" "}
								<kbd className="rounded bg-white px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-gray-800 dark:text-gray-300">
									↑↓
								</kbd>{" "}
								to navigate •
								<kbd className="rounded bg-white px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-gray-800 dark:text-gray-300 mx-1">
									Enter
								</kbd>{" "}
								to open •
								<kbd className="rounded bg-white px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-gray-800 dark:text-gray-300 mx-1">
									ESC
								</kbd>{" "}
								to clear •
								<kbd className="rounded bg-white px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-gray-800 dark:text-gray-300 mx-1">
									/
								</kbd>{" "}
								to focus search
							</p>
						</div>
					</div>
				)}
			</div>

			{showFooter && <Footer />}
		</>
	);
}
