import { OldVersionBanner } from "@/components/OldVersionBanner";

import { VersionInfo, VersionInfoSkeleton } from "@/components/VersionInfo";
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
      <VersionInfo
        currentVersion={version}
        latestVersion={data.versions[0].id}
      />
      {version !== data.versions[0].id && <OldVersionBanner />}
    </div>
  );
}

export function ModuleVersionInfoSkeleton() {
  return <VersionInfoSkeleton />;
}
