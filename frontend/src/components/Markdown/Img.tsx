import { ImgHTMLAttributes } from "react";

export function MarkdownImg({ src, alt }: ImgHTMLAttributes<HTMLImageElement>) {
  return <img src={src} alt={alt} className="inline" />;
}
