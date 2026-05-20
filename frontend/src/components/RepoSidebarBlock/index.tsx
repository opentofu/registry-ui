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
          className="group inline-flex items-center gap-2 rounded-md bg-gray-50 px-3 py-2 text-sm font-medium text-gray-700 transition-all duration-150 hover:bg-gray-100 dark:bg-gray-800/50 dark:text-gray-300 dark:hover:bg-gray-800"
          target="_blank"
          rel="noreferrer noopener"
        >
          <span className="text-gray-500 transition-colors group-hover:text-gray-700 dark:text-gray-400 dark:group-hover:text-gray-300">
            {getLinkLabel(props.link)}
          </span>
          <span className="text-xs text-gray-400 dark:text-gray-500">↗</span>
        </a>
      ) : (
        <span className="flex h-10 w-full animate-pulse rounded-md bg-gray-500/25" />
      )}
    </SidebarBlock>
  );
}

export function RepoSidebarBlockSkeleton() {
  return <RepoSidebarBlock />;
}
