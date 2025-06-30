import { ImgHTMLAttributes } from "react";

interface MarkdownImgProps extends ImgHTMLAttributes<HTMLImageElement> {
  align?: string;
}

export function MarkdownImg({ src, alt, width, align, ...props }: MarkdownImgProps) {
  const alignClass = align === 'right' ? 'float-right ml-4' : align === 'left' ? 'float-left mr-4' : 'inline';
  return <img src={src} alt={alt} width={width} className={alignClass} {...props} />;
}
