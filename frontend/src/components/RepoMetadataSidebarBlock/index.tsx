import {
  MetadataSidebarBlock,
  MetadataSidebarBlockItem,
} from "../MetadataSidebarBlock";
import { github } from "@/icons/github";
import { document } from "@/icons/document";
import { Fragment, ReactNode } from "react";
import { definitions } from "@/api";

function getLinkLabel(url: string) {
  const match = url.match(/github\.com\/([^/]+)\/([^/]+)/);

  if (!match) {
    return null;
  }

  return `${match[1]}/${match[2]}`;
}

interface BlockProps {
  license?: definitions["LicenseList"];
  link?: string | undefined;
}

export function RepoMetadataSidebarBlock(props: BlockProps) {
  let license: ReactNode = (
    <span className="flex h-em w-24 animate-pulse bg-gray-500/25" />
  );

  if (props.license !== undefined) {
    license =
      props.license === null
        ? "Unavailable"
        : props.license.map((license, index, arr) => (
            <Fragment key={license.spdx}>
              <a
                href={license.link}
                key={license.spdx}
                className="underline"
                target="_blank"
                rel="noreferrer noopener"
              >
                {license.spdx}
              </a>
              {index < arr.length - 1 && ", "}
            </Fragment>
          ));
  }

  const link = props.link ? (
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
  );

  return (
    <MetadataSidebarBlock title="Repository">
      <MetadataSidebarBlockItem icon={document} title="License">
        {license}
      </MetadataSidebarBlockItem>
      <MetadataSidebarBlockItem icon={github} title="GitHub">
        {link}
      </MetadataSidebarBlockItem>
    </MetadataSidebarBlock>
  );
}

export function RepoMetadataSidebarBlockSkeleton() {
  return <RepoMetadataSidebarBlock />;
}
