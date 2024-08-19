import { useParams } from "react-router-dom";
import { VersionsSidebarBlock } from "@/components/VersionsSidebarBlock";

import { SidebarBlock } from "@/components/SidebarBlock";
import { getModuleDataQuery } from "../query";
import { useSuspenseQuery } from "@tanstack/react-query";

export function ModuleVersionsSidebarBlock() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  return (
    <VersionsSidebarBlock
      versions={data.versions}
      latestVersion={data.versions[0]}
      currentVersion={version || data.versions[0].id}
      versionLink={(version) =>
        `/module/${namespace}/${name}/${target}/${version}`
      }
    />
  );
}

export function ModuleVersionsSidebarBlockSkeleton() {
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
