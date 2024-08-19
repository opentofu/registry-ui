import {
  MetadataSidebarBlock,
  MetadataSidebarBlockItem,
} from "@/components/MetadataSidebarBlock";
import { github } from "@/icons/github";
import { document } from "@/icons/document";

export function ModuleMetadataSidebarBlock() {
  return (
    <MetadataSidebarBlock title="Repository">
      <MetadataSidebarBlockItem icon={document} title="License">
        <a href="https://opentofu.org" className="underline">
          Unknown
        </a>
      </MetadataSidebarBlockItem>
      <MetadataSidebarBlockItem icon={github} title="GitHub">
        <a href="https://opentofu.org" className="underline">
          Unknown
        </a>
      </MetadataSidebarBlockItem>
    </MetadataSidebarBlock>
  );
}
