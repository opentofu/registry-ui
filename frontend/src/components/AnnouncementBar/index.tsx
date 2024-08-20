import * as prod from "react/jsx-runtime";
import { useMemo } from "react";
import { unified } from "unified";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeReact, { Options } from "rehype-react";

const production: Options = {
  development: false,
  Fragment: prod.Fragment,
  jsx: prod.jsx,
  jsxs: prod.jsxs,
};

interface AnnouncementBarProps {
  content: string;
}

export function AnnouncementBar({ content }: AnnouncementBarProps) {
  const { result } = useMemo(
    () =>
      unified()
        .use(remarkParse)
        .use(remarkRehype)
        .use(rehypeReact, production)
        .processSync(content),
    [content],
  );

  if (!result) {
    return null;
  }

  return (
    <div
      className="flex h-6 items-center justify-center bg-brand-600 text-sm text-gray-900 [&_a]:underline"
      role="banner"
    >
      {result}
    </div>
  );
}
