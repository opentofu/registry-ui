import { Header } from "../Header";
import { Breadcrumbs } from "../Breadcrumbs";
import { ReactNode } from "react";
import { Footer } from "../Footer";
import PatternBg from "../PatternBg";

interface SimpleLayoutProps {
  children: ReactNode;
}

export function SimpleLayout({ children }: SimpleLayoutProps) {
  return (
    <>
      <PatternBg />
      <div className="fixed inset-0 -z-10 bg-white/50 dark:bg-blue-950/50" />
      <Header />
      <div className="mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 pt-24">
        <div className="h-10 flex items-center px-3">
          <Breadcrumbs className="h-10 flex-1" />
        </div>
        <main className="pb-5">
          {children}
        </main>
      </div>
      <Footer />
    </>
  );
}
