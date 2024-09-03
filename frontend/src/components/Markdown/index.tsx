import { useMemo } from "react";
import * as prod from "react/jsx-runtime";

import remarkFrontmatter from "remark-frontmatter";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeSlug from "rehype-slug";
import remarkGfm from "remark-gfm";
import rehypeReact, { Options } from "rehype-react";
import rehypeSanitize from "rehype-sanitize";
import rehypeRaw from "rehype-raw";
import { unified } from "unified";
import { MarkdownH1 } from "./H1";
import { MarkdownP } from "./P";
import { MarkdownPre } from "./Pre";
import { MarkdownH2 } from "./H2";
import { MarkdownH3 } from "./H3";
import { MarkdownCode } from "./Code";
import { MarkdownUl } from "./Ul";
import { MarkdownLi } from "./Li";
import { MarkdownA } from "./A";
import { MarkdownTable } from "./Table";
import { MarkdownTd } from "./Td";
import { MarkdownTh } from "./Th";
import { MarkdownImg } from "./Img";
import { MarkdownOl } from "./Ol";
import { MarkdownHr } from "./Hr";

const production: Options = {
  development: false,
  Fragment: prod.Fragment,
  jsx: prod.jsx,
  jsxs: prod.jsxs,
  components: {
    a: MarkdownA,
    ul: MarkdownUl,
    li: MarkdownLi,
    h1: MarkdownH1,
    h2: MarkdownH2,
    h3: MarkdownH3,
    p: MarkdownP,
    code: MarkdownCode,
    pre: MarkdownPre,
    table: MarkdownTable,
    td: MarkdownTd,
    th: MarkdownTh,
    img: MarkdownImg,
    ol: MarkdownOl,
    hr: MarkdownHr,
  },
};

interface MarkdownProps {
  text: string;
}

export function Markdown({ text }: MarkdownProps) {
  const { result } = useMemo(
    () =>
      unified()
        .use(remarkParse)
        .use(remarkFrontmatter)
        .use(remarkGfm)
        .use(remarkRehype, { allowDangerousHtml: true }) // This is okay to use dangerous html because we are sanitizing later on in the pipeline
        .use(rehypeRaw)
        .use(rehypeSanitize)
        .use(rehypeSlug)
        .use(rehypeReact, production)
        .processSync(text),
    [text],
  );

  return result;
}
