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
      <main className="container mx-auto flex grow flex-col">
        <Breadcrumbs />
        {children}
      </main>
    </>
  );
}
