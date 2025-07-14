import { useMemo } from "react";
import { processor } from "./processor";

interface MarkdownProps {
  text: string;
  onTocExtracted?: (toc: TocEntry[]) => void;
}

export interface TocEntry {
  value: string;
  depth: number;
  id?: string;
  children?: TocEntry[];
}

export function Markdown({ text, onTocExtracted }: MarkdownProps) {
  const { result, toc } = useMemo(() => {
    const file = processor.processSync(text.trimStart());
    const toc = (file.data.toc as TocEntry[]) || [];
    return { result: file.result, toc };
  }, [text]);

  // Call the callback with TOC data when it changes
  useMemo(() => {
    if (onTocExtracted) {
      onTocExtracted(toc);
    }
  }, [toc, onTocExtracted]);

  return result;
}
