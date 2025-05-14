import { useMemo } from "react";
import { processor } from "./processor";

interface MarkdownProps {
  text: string;
}

export function Markdown({ text }: MarkdownProps) {
  const { result } = useMemo(
    () => processor.processSync(text.trimStart()),
    [text],
  );

  return result;
}
