import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";

import {
  RepoSidebarBlock,
  RepoSidebarBlockSkeleton,
} from "@/components/RepoSidebarBlock";
import { LicenseSidebarBlock } from "@/components/LicenseSidebarBlock";

export function ProviderMetadataSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  return (
    <>
      <LicenseSidebarBlock license={data.license} />
      <RepoSidebarBlock link={data.link} />
    </>
  );
}

export function ProviderMetadataSidebarBlockSkeleton() {
  return (
    <>
      <LicenseSidebarBlock />
      <RepoSidebarBlockSkeleton />
    </>
  );
}
