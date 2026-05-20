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
            <Icon path={github} className="size-5 shrink-0" />
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
          className="inline-flex items-center gap-2 px-3 py-2 rounded-md bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 transition-all duration-150 text-sm font-medium text-gray-700 dark:text-gray-300 group"
          target="_blank"
          rel="noreferrer noopener"
        >
          <span className="text-gray-500 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-300 transition-colors">
            {getLinkLabel(props.link)}
          </span>
          <span className="text-xs text-gray-400 dark:text-gray-500">↗</span>
        </a>
      ) : (
        <span className="flex h-10 w-full animate-pulse bg-gray-500/25 rounded-md" />
      )}
    </SidebarBlock>
  );
}

export function RepoSidebarBlockSkeleton() {
  return <RepoSidebarBlock />;
}
