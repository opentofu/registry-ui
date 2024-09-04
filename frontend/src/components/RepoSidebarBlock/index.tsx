import { SidebarBlock } from "../SidebarBlock";

function getLinkLabel(url: string) {
  const match = url.match(/github\.com\/([^/]+)\/([^/]+)/);

  if (!match) {
    return null;
  }

  return `${match[1]}/${match[2]}`;
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
          className="underline"
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
