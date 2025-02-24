import { Footer } from "../Footer";
import { Header } from "../Header";
import { ReactNode } from "react";

interface SidebarLayoutProps {
  children: ReactNode;
  before?: ReactNode;
  after?: ReactNode;
}

export function SidebarLayout({ children, before, after }: SidebarLayoutProps) {
  return (
    <>
      <Header />
      <div className="max-w-8xl mx-auto flex w-full grow divide-x divide-gray-200 px-5 dark:divide-gray-800">
        {before}
        <main className="min-w-0 flex-1">{children}</main>
        {after}
      </div>
      <Footer />
    </>
  );
}
