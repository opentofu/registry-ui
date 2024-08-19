import { VersionsSidebarBlock } from "@/components/VersionsSidebarBlock";

import { SidebarBlock } from "@/components/SidebarBlock";
import { getProviderDataQuery } from "../query";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useProviderParams } from "../hooks/useProviderParams";

export function ProviderVersionsSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(getProviderDataQuery(namespace, provider));

  return (
    <VersionsSidebarBlock
      versions={data.versions}
      latestVersion={data.versions[0]}
      currentVersion={version || data.versions[0].id}
      versionLink={(version) => `/provider/${namespace}/${provider}/${version}`}
    />
  );
}

export function ProviderVersionsSidebarBlockSkeleton() {
  return (
    <SidebarBlock title="Versions">
      <span className="mt-6 flex animate-pulse items-center justify-between">
        <span className="flex h-em w-16 bg-gray-500/25" />
        <span className="flex h-em w-32 bg-gray-500/25" />
      </span>
      <span className="mt-5 flex animate-pulse items-center justify-between">
        <span className="flex h-em w-20 bg-gray-500/25" />
        <span className="flex h-em w-28 bg-gray-500/25" />
      </span>
      <span className="mt-5 flex animate-pulse items-center justify-between">
        <span className="flex h-em w-12 bg-gray-500/25" />
        <span className="flex h-em w-36 bg-gray-500/25" />
      </span>
    </SidebarBlock>
  );
}
