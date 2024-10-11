import { matter } from "vfile-matter";
import { Document, SidebarItem } from "./types";
import sidebar from "../../../docs/sidebar.json";
import { processor } from "@/components/Markdown/processor";
import { renderToStaticMarkup } from "react-dom/server";

export function getSlugPathMap() {
  const docs = import.meta.glob("../../../docs/**/*.md", {
    eager: true,
    as: "raw",
  });

  const processorWithFrontmatter = processor().use(() => {
    return (_, file) => {
      matter(file);
    };
  });

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
