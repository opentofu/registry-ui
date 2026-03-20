import { createContext, useContext, useState, ReactNode } from "react";
import { TocEntry } from "@/components/Markdown";

interface DocsContextValue {
  toc: TocEntry[];
  setToc: (toc: TocEntry[]) => void;
}

const DocsContext = createContext<DocsContextValue | undefined>(undefined);

export function DocsProvider({ children }: { children: ReactNode }) {
  const [toc, setToc] = useState<TocEntry[]>([]);

  return (
    <DocsContext.Provider value={{ toc, setToc }}>
      {children}
    </DocsContext.Provider>
  );
}

export function useDocsContext() {
  const context = useContext(DocsContext);
  if (!context) {
    throw new Error("useDocsContext must be used within a DocsProvider");
  }
  return context;
}