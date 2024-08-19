import { HTMLAttributes } from "react";

export function MarkdownImg({ src, alt }: HTMLAttributes<HTMLImageElement>) {
  return <img src={src} alt={alt} className="inline" />;
}
