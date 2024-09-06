import { defer, LoaderFunction } from "react-router-dom";
import { NotFoundPageError } from "@/utils/errors";
import sidebar from "../../../docs/sidebar.json";

interface Document {
  content: string;
  frontmatter: {
    title: string;
  };
}

const docs = import.meta.glob<true, "md", Document>("../../../docs/**/*.md", {
  eager: true,
});

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

function buildSlugPathMap(sidebarItems: SidebarItem[]) {
  const slugPathMap: Record<string, string> = {};

  const traverseItems = (items: SidebarItem[]) => {
    for (const item of items) {
      if (item.items) {
        traverseItems(item.items);
      } else {
        slugPathMap[item.slug] = item.path;
      }
    }
  };

  traverseItems(sidebarItems);

  return slugPathMap;
}

export const docsLoader: LoaderFunction = async ({ params }) => {
  const { "*": slug = "" } = params;
  const normalizedSlug = slug.replace(/[^a-zA-Z0-9/-]/g, "");

  const slugPathMap = buildSlugPathMap(sidebar);
  const sidebarItem = slugPathMap[normalizedSlug];

  if (!sidebarItem) {
    throw new NotFoundPageError();
  }

  const document = docs[`../../../docs/${sidebarItem}`];

  if (!document) {
    throw new NotFoundPageError();
  }

  return defer({
    content: document.content,
    frontmatter: document.frontmatter,
  });
};
