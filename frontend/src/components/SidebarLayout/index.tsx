import { Footer } from "../Footer";
import { Header } from "../Header";
import { ReactNode } from "react";
import PatternBg from "../PatternBg";
import { Breadcrumbs } from "../Breadcrumbs";

interface SidebarLayoutProps {
  children: ReactNode;
  before?: ReactNode;
  after?: ReactNode;
  showBreadcrumbs?: boolean;
}

export function SidebarLayout({ children, before, after, showBreadcrumbs = false }: SidebarLayoutProps) {
  return (
    <>
      <PatternBg />
      <div className="fixed inset-0 -z-10 bg-white/50 dark:bg-blue-950/50" />
      <Header />
      <div className="mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 pt-24">
        <div className="h-10 bg-gray-200 dark:bg-blue-950 border border-gray-300 dark:border-gray-700 border-b-0 flex items-center px-3 rounded-t">
          {showBreadcrumbs ? (
            <Breadcrumbs className="h-10 flex-1" />
          ) : (
            <span className="font-mono text-sm text-gray-600 dark:text-gray-400">◆ opentofu-registry</span>
          )}
        </div>
        <div className="flex flex-1 divide-x divide-gray-200 dark:divide-gray-800 border border-gray-300 dark:border-gray-700">
          {before}
          <main className="min-w-0 flex-1 bg-gray-100 dark:bg-blue-900">
            <div className="mt-8">
              {children}
            </div>
          </main>
          {after}
        </div>
      </div>
      <Footer />
    </>
  );
}
