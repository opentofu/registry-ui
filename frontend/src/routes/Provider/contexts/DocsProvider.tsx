import { ReactNode, useState } from "react";
import { TocEntry } from "@/components/Markdown";
import { DocsContext } from "./DocsContext";

export function DocsProvider({ children }: { children: ReactNode }) {
  const [toc, setToc] = useState<TocEntry[]>([]);

  return (
    <DocsContext.Provider value={{ toc, setToc }}>
      {children}
    </DocsContext.Provider>
  );
}
