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
      <main className="mx-auto flex w-full max-w-(--breakpoint-3xl) grow flex-col px-5 pb-5 pt-24">
        <Breadcrumbs />
        {children}
      </main>
      <Footer />
    </>
  );
}
