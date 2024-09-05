import { github } from "@/icons/github";
import { Icon } from "../Icon";
import { SidebarBlock } from "../SidebarBlock";

function getLinkLabel(url: string) {
  try {
    const parsedUrl = new URL(url);

    switch (parsedUrl.hostname) {
      case "github.com": {
        const pathParts = parsedUrl.pathname.split("/");

        return (
          <>
            <Icon path={github} className="mt-1.5 size-em shrink-0" />
            <span>
              {pathParts[1]}/{pathParts[2]}
            </span>
          </>
        );
      }
      default:
        return `${parsedUrl.hostname}${parsedUrl.pathname}`;
    }
  } catch {
    return url;
  }
}

interface BlockProps {
  link?: string | undefined;
}

export function RepoSidebarBlock(props: BlockProps) {
  return (
    <SidebarBlock title="Repository">
      {props.link ? (
        <a
          href={props.link}
          className="inline-flex gap-2 underline"
          target="_blank"
          rel="noreferrer noopener"
        >
          {getLinkLabel(props.link)}
        </a>
      ) : (
        <span className="flex h-em w-32 animate-pulse bg-gray-500/25" />
      )}
    </SidebarBlock>
  );
}

export function RepoSidebarBlockSkeleton() {
  return <RepoSidebarBlock />;
}
