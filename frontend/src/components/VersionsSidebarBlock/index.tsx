import { definitions } from "@/api";
import clsx from "clsx";
import { useMemo, useState } from "react";
import { NavLink } from "react-router";
import { chevron } from "@/icons/chevron";
import { expand } from "@/icons/expand";
import { DateTime } from "../DateTime";
import { Icon } from "../Icon";
import { SidebarBlock } from "../SidebarBlock";
import { TreeView, TreeViewItem } from "../TreeView";
import { groupVersions } from "./utils";

interface SimpleTreeNode {
  label: string;
  published: string;
  isActive: boolean;
  link: string;
}

interface TreeNode extends SimpleTreeNode {
  children?: Array<SimpleTreeNode>;
}

interface TreeItemProps {
  node: TreeNode;
}

function VersionTreeViewItemHandle({
  published,
  label,
  isActive,
  isOpen,
  onClick,
  className,
  link,
}: {
  published: string;
  label: string;
  isActive: boolean;
  isOpen?: boolean;
  onClick?: () => void;
  className?: string;
  link?: string;
}) {
  let handle;

  if (onClick) {
    handle = (
      <button
        className={clsx(
          "inline-flex items-center gap-2 text-sm font-medium transition-colors duration-150",
          isActive
            ? "text-brand-800 dark:text-brand-200"
            : "text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white",
        )}
        onClick={() => onClick()}
      >
        <Icon
          path={chevron}
          className={clsx(
            "size-3.5 transition-transform duration-200",
            isOpen && "rotate-90",
          )}
        />
        {label}
      </button>
    );
  } else if (link) {
    handle = (
      <NavLink
        to={link}
        className={clsx(
          "text-sm transition-colors duration-150",
          isActive
            ? "text-brand-800 dark:text-brand-200 font-medium"
            : "text-gray-700 underline-offset-2 hover:text-gray-900 hover:underline dark:text-gray-300 dark:hover:text-white",
        )}
        state={{ fromVersion: true }}
      >
        {label}
      </NavLink>
    );
  }

  return (
    <span
      className={clsx(
        "flex w-full items-center justify-between rounded-md px-3 py-2 transition-all duration-150 hover:bg-gray-100 dark:hover:bg-gray-800",
        isActive && "bg-brand-300/40 dark:bg-brand-300/40",
        className,
      )}
    >
      {handle}
      <span className="text-xs text-gray-500 dark:text-gray-400">
        <DateTime value={published} />
      </span>
    </span>
  );
}

function VersionTreeViewNestedItems({
  children,
}: {
  children: SimpleTreeNode[];
}) {
  const [expanded, setExpanded] = useState(false);
  const activeIndex = children.findIndex((child) => child.isActive);
  const visibleCount = expanded ? children.length : 5;

  const visibleChildren = useMemo(() => {
    const start = Math.max(
      0,
      Math.min(activeIndex - 2, children.length - visibleCount),
    );

    const end = Math.min(start + visibleCount, children.length);
    return children.slice(start, end);
  }, [activeIndex, children, visibleCount]);

  return (
    <TreeView className="mt-1 ml-4">
      {visibleChildren.map((node, index) => (
        <TreeViewItem nested key={index}>
          <VersionTreeViewItemHandle
            published={node.published}
            label={node.label}
            isActive={node.isActive}
            className=""
            link={node.link}
          />
        </TreeViewItem>
      ))}

      {(children.length > visibleChildren.length || expanded) && (
        <TreeViewItem nested>
          <button
            onClick={() => setExpanded((v) => !v)}
            className="inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm text-gray-600 transition-colors duration-150 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white"
          >
            <Icon path={expand} className="size-4" />
            {expanded ? "Show less" : "Show more"}
          </button>
        </TreeViewItem>
      )}
    </TreeView>
  );
}

function VersionTreeViewItem({ node }: TreeItemProps) {
  const [isOpen, setIsOpen] = useState(node.isActive);

  return (
    <TreeViewItem>
      <VersionTreeViewItemHandle
        published={node.published}
        label={node.label}
        isActive={node.isActive}
        isOpen={isOpen}
        onClick={node.children ? () => setIsOpen(!isOpen) : undefined}
        link={node.link}
      />

      {isOpen && <VersionTreeViewNestedItems children={node.children || []} />}
    </TreeViewItem>
  );
}

interface VersionsSidebarBlockProps {
  versions: Array<
    | definitions["ProviderVersionDescriptor"]
    | definitions["ModuleVersionDescriptor"]
  >;
  latestVersion:
  | definitions["ProviderVersionDescriptor"]
  | definitions["ModuleVersionDescriptor"];
  currentVersion: string;
  versionLink: (version: string) => string;
}

export function VersionsSidebarBlock({
  versions,
  latestVersion,
  currentVersion,
  versionLink,
}: VersionsSidebarBlockProps) {
  const groupedVersions = groupVersions({
    versions,
    currentVersion,
    latestVersion: latestVersion.id,
    versionLink,
  });

  return (
    <SidebarBlock title="Versions">
      <div className="space-y-4">
        <div>
          <h4 className="mb-2 text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400">
            Latest
          </h4>
          <a
            href={versionLink("latest")}
            className={clsx(
              "flex items-center justify-between rounded-md p-3 transition-all duration-150",
              currentVersion === latestVersion.id
                ? "bg-brand-300/40 dark:bg-brand-300/40"
                : "hover:bg-gray-100 dark:hover:bg-gray-800",
            )}
          >
            <div className="flex items-center gap-2">
              <span
                className={clsx(
                  "text-sm font-medium",
                  currentVersion === latestVersion.id
                    ? "text-brand-800 dark:text-brand-200"
                    : "text-gray-700 dark:text-gray-300",
                )}
              >
                {latestVersion.id}
              </span>
              <span className="rounded-full bg-gray-100 px-1.5 py-0.5 text-xs font-medium text-gray-700 dark:bg-gray-800 dark:text-gray-300">
                Latest
              </span>
            </div>
            <span className="text-xs text-gray-500 dark:text-gray-400">
              <DateTime value={latestVersion.published} />
            </span>
          </a>
        </div>

        {groupedVersions.length > 0 && (
          <div>
            <h4 className="mb-2 text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400">
              Older versions
            </h4>
            <TreeView>
              {groupedVersions.map((node) => (
                <VersionTreeViewItem key={node.label} node={node} />
              ))}
            </TreeView>
          </div>
        )}
      </div>
    </SidebarBlock>
  );
}
