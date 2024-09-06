import { Plugin } from "unified";
import { matter } from "vfile-matter";
import { Document } from "./types";
import sidebar from "../../../docs/sidebar.json";
import { processor } from "@/components/Markdown/processor";
import { renderToStaticMarkup } from "react-dom/server";

const docs = import.meta.glob("../../../docs/**/*.md", {
  eager: true,
  as: "raw",
});

const extractFrontmatter: Plugin = () => {
  return (_, file) => {
    matter(file);
  };
};

const processorWithFrontmatter = processor().use(extractFrontmatter);

const documents = Object.fromEntries(
  Object.entries(docs).map(([path, document]) => {
    const { data, result } = processorWithFrontmatter.processSync(document);

    return [
      path,
      {
        data: data.matter,
        content: renderToStaticMarkup(result),
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
