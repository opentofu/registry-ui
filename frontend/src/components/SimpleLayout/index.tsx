import { Header } from "../Header";
import { Breadcrumbs } from "../Breadcrumbs";
import { ReactNode } from "react";

interface SimpleLayoutProps {
  children: ReactNode;
}

export function SimpleLayout({ children }: SimpleLayoutProps) {
  return (
    <>
      <Header />
      <main className="mx-auto flex w-full max-w-screen-3xl grow flex-col px-5">
        <Breadcrumbs />
        {children}
      </main>
    </>
  );
}
