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

  const latestVersion = data.versions[0].id;

  const latestVersionLink = `/module/${namespace}/${name}/${target}/${latestVersion}`;

  return (
    <div className="flex flex-col gap-5">
      <VersionInfo currentVersion={version} latestVersion={latestVersion} />
      {version !== latestVersion && (
        <OldVersionBanner latestVersionLink={latestVersionLink} />
      )}
    </div>
  );
}

export function ModuleVersionInfoSkeleton() {
  return <VersionInfoSkeleton />;
}
