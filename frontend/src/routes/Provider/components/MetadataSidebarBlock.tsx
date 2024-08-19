import {
  MetadataSidebarBlock,
  MetadataSidebarBlockItem,
} from "@/components/MetadataSidebarBlock";

import { github } from "@/icons/github";
import { document } from "@/icons/document";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";

function Block(props: { licenses?: string[]; repository?: string }) {
  return (
    <MetadataSidebarBlock title="Repository">
      <MetadataSidebarBlockItem icon={document} title="License">
        {props.licenses ? (
          props.licenses.map((license) => (
            <span key={license} className="mr-2">
              {license}
            </span>
          ))
        ) : (
          <span className="flex h-em w-24 animate-pulse bg-gray-500/25" />
        )}
      </MetadataSidebarBlockItem>
      <MetadataSidebarBlockItem icon={github} title="GitHub">
        {props.repository ? (
          <a href="https://opentofu.org" className="underline">
            {props.repository}
          </a>
        ) : (
          <span className="flex h-em w-32 animate-pulse bg-gray-500/25" />
        )}
      </MetadataSidebarBlockItem>
    </MetadataSidebarBlock>
  );
}

export function ProviderMetadataSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  const licenses = data.license.map((license) => license.spdx);

  return <Block licenses={licenses} />;
}

export function ProviderMetadataSidebarBlockSkeleton() {
  return <Block />;
}
