import { NavLink } from "react-router-dom";
import { SidebarBlock } from "../SidebarBlock";
import { TreeView, TreeViewItem } from "../TreeView";
import { useMemo, useState } from "react";
import clsx from "clsx";
import { Icon } from "../Icon";
import { chevron } from "../../icons/chevron";
import { groupVersions } from "./utils";
import { expand } from "../../icons/expand";
import { formatDate } from "../../utils/formatDate";
import { definitions } from "@/api";

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
          "inline-flex items-center gap-2",
          isActive ? "text-brand-700 dark:text-brand-600" : "text-inherit",
        )}
        onClick={() => onClick()}
      >
        <Icon
          path={chevron}
          className={clsx("size-4 text-inherit", isOpen && "rotate-90")}
        />

        {label}
      </button>
    );
  } else if (link) {
    handle = (
      <NavLink
        to={link}
        className={
          isActive
            ? "text-brand-700 dark:text-brand-600"
            : "text-inherit underline underline-offset-2"
        }
      >
        {label}
      </NavLink>
    );
  }

  return (
    <span
      className={clsx(
        "flex w-full items-center justify-between py-2",
        className,
      )}
    >
      {handle}
      <span className="text-gray-700 dark:text-gray-300">
        {formatDate(published)}
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
    const start = Math.max(0, activeIndex - Math.floor(visibleCount / 2));
    const end = Math.min(start + visibleCount, children.length);
    return children.slice(start, end);
  }, [activeIndex, children, visibleCount]);

  return (
    <TreeView className="ml-2">
      {visibleChildren.map((node, index) => (
        <TreeViewItem nested key={index}>
          <VersionTreeViewItemHandle
            published={node.published}
            label={node.label}
            isActive={node.isActive}
            className="pl-4"
            link={node.link}
          />
        </TreeViewItem>
      ))}

      {(children.length > visibleChildren.length || expanded) && (
        <TreeViewItem nested>
          <button
            onClick={() => setExpanded((v) => !v)}
            className="inline-flex items-center gap-2 px-4 py-2 text-gray-700 dark:text-gray-300"
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
      <TreeView>
        <VersionTreeViewItem
          node={{
            label: `${latestVersion.id} (latest)`,
            published: latestVersion.published,
            isActive: currentVersion === latestVersion.id,
            link: versionLink("latest"),
          }}
        />

        {groupedVersions.map((node) => (
          <VersionTreeViewItem key={node.label} node={node} />
        ))}
      </TreeView>
    </SidebarBlock>
  );
}
