import * as prod from "react/jsx-runtime";

import remarkFrontmatter from "remark-frontmatter";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeSlug from "rehype-slug";
import remarkGfm from "remark-gfm";
import remarkGithubAlerts from "remark-github-alerts";
import rehypeReact, { Options } from "rehype-react";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import rehypeRaw from "rehype-raw";
import { unified } from "unified";
import { rehypeCodeListAnchors } from "./plugins/rehypeCodeListAnchors";
import rehypeExtractToc from "@stefanprobst/rehype-extract-toc";

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

const rehypeReactOptions: Options = {
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

const sanitizeSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    img: [...(defaultSchema.attributes?.img || []), 'align', 'width', 'height'],
    div: [...(defaultSchema.attributes?.div || []), 'className', 'class'],
    p: [...(defaultSchema.attributes?.p || []), 'className', 'class', 'align'],
    svg: [...(defaultSchema.attributes?.svg || []), 'className', 'class', 'viewBox', 'fill', 'height', 'width', 'style'],
    path: [...(defaultSchema.attributes?.path || []), 'd', 'fillRule', 'clipRule'],

    li: [...(defaultSchema.attributes?.li || []), 'id'],
    a: [...(defaultSchema.attributes?.a || []), 'href', 'className'],
    span: [...(defaultSchema.attributes?.span || []), 'id', 'className'],
  },
  tagNames: [...(defaultSchema.tagNames || []), 'svg', 'path']
};

export const processor = unified()
  .use(remarkParse)
  .use(remarkFrontmatter)
  .use(remarkGfm)
  .use(remarkGithubAlerts)
  .use(remarkRehype, {
    // This is okay to use dangerous html because we are sanitizing later on in the pipeline
    allowDangerousHtml: true,
  })
  .use(rehypeRaw)
  .use(rehypeSanitize, sanitizeSchema)
  .use(rehypeSlug)
  .use(rehypeExtractToc)
  .use(rehypeCodeListAnchors)
  .use(rehypeReact, rehypeReactOptions);
