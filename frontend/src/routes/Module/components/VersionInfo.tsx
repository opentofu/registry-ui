import { OldVersionBanner } from "@/components/OldVersionBanner";

import { VersionInfo } from "@/components/VersionInfo";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";

export function ModuleVersionInfo() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  return (
    <div className="flex flex-col gap-5">
      <div className="flex items-center justify-between">
        <VersionInfo
          currentVersion={version}
          latestVersion={data.versions[0].id}
        />
      </div>
      {version !== data.versions[0].id && <OldVersionBanner />}
    </div>
  );
}

export function ModuleVersionInfoSkeleton() {
  return <span className="flex h-em w-48 animate-pulse bg-gray-500/25" />;
}
