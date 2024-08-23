import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";

import {
  RepoMetadataSidebarBlock,
  RepoMetadataSidebarBlockSkeleton,
} from "@/components/RepoMetadataSidebarBlock";

export function ProviderMetadataSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  return <RepoMetadataSidebarBlock license={data.license} link={data.link} />;
}

export { RepoMetadataSidebarBlockSkeleton as ProviderMetadataSidebarBlockSkeleton };
