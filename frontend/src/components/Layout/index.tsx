import { ReactNode } from "react";
import PatternBg from "../PatternBg";
import { Header } from "../Header";
import { Footer } from "../Footer";

interface LayoutProps {
  children: ReactNode;
  showFooter?: boolean;
}

export function Layout({ children, showFooter = true }: LayoutProps) {
  return (
    <>
      <PatternBg />
      <Header />
      {children}
      {showFooter && <Footer />}
    </>
  );
}