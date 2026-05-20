import { createContext, useContext } from "react";
import { TocEntry } from "@/components/Markdown";

interface DocsContextValue {
  toc: TocEntry[];
  setToc: (toc: TocEntry[]) => void;
}

export const DocsContext = createContext<DocsContextValue | undefined>(
  undefined,
);

export function useDocsContext() {
  const context = useContext(DocsContext);
  if (!context) {
    throw new Error("useDocsContext must be used within a DocsProvider");
  }
  return context;
}
