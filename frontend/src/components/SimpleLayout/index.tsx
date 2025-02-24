import { Header } from "../Header";
import { Breadcrumbs } from "../Breadcrumbs";
import { ReactNode } from "react";
import { Footer } from "../Footer";

interface SimpleLayoutProps {
  children: ReactNode;
}

export function SimpleLayout({ children }: SimpleLayoutProps) {
  return (
    <>
      <Header />
      <main className="max-w max-w-8xl mx-auto flex w-full grow flex-col px-5 pb-5">
        <Breadcrumbs />
        {children}
      </main>
      <Footer />
    </>
  );
}
