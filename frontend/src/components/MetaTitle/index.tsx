import { useEffect } from "react";

interface MetaTitleProps {
  children: string;
}

export function MetaTitle({ children }: MetaTitleProps) {
  useEffect(() => {
    document.title = `${children} - OpenTofu Registry`;

    return () => {
      document.title = "OpenTofu Registry";
    };
  }, [children]);

  return null;
}
