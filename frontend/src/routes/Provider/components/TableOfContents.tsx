import { useDocsContext } from "../contexts/DocsContext";
import { TocEntry } from "@/components/Markdown";
import { Link } from "react-router";
import { useLocation } from "react-router";
import { useScrollToAnchor } from "@/hooks/useScrollToAnchor";

export function TableOfContents() {
  const { toc } = useDocsContext();
  const location = useLocation();
  const scrollToAnchor = useScrollToAnchor();

  // Get the children of H1 headings (which are the H2+ headings we want to show)
  const getVisibleTocItems = (items: TocEntry[]): TocEntry[] => {
    // If the top level has H1s, return their children
    if (items.length > 0 && items[0].depth === 1) {
      return items.flatMap(item => item.children || []);
    }
    // Otherwise return the items as-is
    return items;
  };

  const visibleToc = getVisibleTocItems(toc);

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    e.preventDefault();
    scrollToAnchor(id);
  };

  if (visibleToc.length === 0) {
    return null;
  }

  const renderTocItem = (item: TocEntry, index: number) => {
    const isActive = location.hash === `#${item.id}`;
    const depth = Math.min(item.depth - 2, 3); // Adjust depth since we filter H1, cap at h4 level
    const paddingLeft = `${depth * 0.75}rem`;

    return (
      <li key={`${item.id}-${index}`}>
        <Link
          to={`#${item.id}`}
          className={`block py-1 text-sm transition-colors hover:text-primary ${
            isActive
              ? "font-medium text-primary"
              : "text-gray-600 dark:text-gray-400"
          }`}
          style={{ paddingLeft }}
          onClick={(e) => handleClick(e, item.id || "")}
        >
          {item.value}
        </Link>
        {item.children && item.children.length > 0 && (
          <ul>
            {item.children.map((child, childIndex) =>
              renderTocItem(child, childIndex)
            )}
          </ul>
        )}
      </li>
    );
  };

  return (
    <div className="space-y-3 px-5 py-4">
      <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
        On this page
      </h3>
      <nav>
        <ul className="space-y-1">
          {visibleToc.map((item, index) => renderTocItem(item, index))}
        </ul>
      </nav>
    </div>
  );
}

export function TableOfContentsSkeleton() {
  return (
    <div className="space-y-3 px-5 py-4">
      <span className="flex h-4 w-20 animate-pulse bg-gray-500/25" />
      <div className="space-y-2">
        <span className="flex h-3 w-32 animate-pulse bg-gray-500/25" />
        <span className="ml-3 flex h-3 w-28 animate-pulse bg-gray-500/25" />
        <span className="ml-3 flex h-3 w-24 animate-pulse bg-gray-500/25" />
        <span className="flex h-3 w-30 animate-pulse bg-gray-500/25" />
        <span className="ml-3 flex h-3 w-26 animate-pulse bg-gray-500/25" />
      </div>
    </div>
  );
}