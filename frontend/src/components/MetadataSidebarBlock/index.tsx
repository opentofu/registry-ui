import { ReactNode } from "react";
import { SidebarBlock } from "../SidebarBlock";
import { Icon } from "../Icon";

interface MetadataSidebarBlockProps {
  children: ReactNode;
  title: string;
}

export function MetadataSidebarBlock({
  children,
  title,
}: MetadataSidebarBlockProps) {
  return (
    <SidebarBlock title={title}>
      <dl className="mt-4 flex flex-col gap-4">{children}</dl>
    </SidebarBlock>
  );
}

interface MetadataSidebarBlockItemProps {
  title: string;
  children: ReactNode;
  icon: string;
}

export function MetadataSidebarBlockItem({
  icon,
  title,
  children,
}: MetadataSidebarBlockItemProps) {
  return (
    <div className="flex justify-between">
      <dt className="flex items-start gap-2 text-gray-700 dark:text-gray-300">
        <Icon path={icon} className="size-5" />
        {title}
      </dt>
      <dd className="flex flex-col justify-center text-right">{children}</dd>
    </div>
  );
}
