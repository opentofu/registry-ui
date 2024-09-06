import { Plugin, unified } from "unified";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeStringify from "rehype-stringify";
import remarkFrontmatter from "remark-frontmatter";
import { matter } from "vfile-matter";
import { Document } from "./types";

import sidebar from "../../../docs/sidebar.json";

const docs = import.meta.glob("../../../docs/**/*.md", {
  eager: true,
  as: "raw",
});

const extractFrontmatter: Plugin = () => {
  return (_, file) => {
    matter(file);
  };
};

const processor = unified()
  .use(remarkParse)
  .use(remarkFrontmatter)
  .use(extractFrontmatter)
  .use(remarkRehype)
  .use(rehypeStringify);

const documents = Object.fromEntries(
  Object.entries(docs).map(([path, document]) => {
    const { data, value } = processor.processSync(document);
    return [
      path,
      {
        data: data.matter,
        content: value,
      },
    ];
  }),
);

type SidebarItem =
  | {
      title: string;
      items: SidebarItem[];
    }
  | {
      title: string;
      slug: string;
      path: string;
      items?: never;
    };

function getSlugPathMap() {
  const slugPathMap: Record<string, Document> = {};

  const traverseItems = (items: SidebarItem[]) => {
    for (const item of items) {
      if (item.items) {
        traverseItems(item.items);
      } else {
        slugPathMap[item.slug] = documents[`../../../docs/${item.path}`];
      }
    }
  };

  traverseItems(sidebar);

  return slugPathMap;
}

export const slugPathMap = getSlugPathMap();
