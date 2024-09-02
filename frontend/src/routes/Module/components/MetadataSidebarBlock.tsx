import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";

import {
  RepoSidebarBlock,
  RepoSidebarBlockSkeleton,
} from "@/components/RepoSidebarBlock";
import { LicenseSidebarBlock } from "@/components/LicenseSidebarBlock";

export function ModuleMetadataSidebarBlock() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <>
      <LicenseSidebarBlock license={data.licenses} />
      <RepoSidebarBlock link={data.link} />
    </>
  );
}

export function ModuleMetadataSidebarBlockSkeleton() {
  return (
    <>
      <LicenseSidebarBlock />
      <RepoSidebarBlockSkeleton />
    </>
  );
}
