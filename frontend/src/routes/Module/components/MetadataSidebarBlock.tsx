import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";

import {
  RepoMetadataSidebarBlock,
  RepoMetadataSidebarBlockSkeleton,
} from "@/components/RepoMetadataSidebarBlock";

export function ModuleMetadataSidebarBlock() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return <RepoMetadataSidebarBlock license={data.licenses} link={data.link} />;
}

export { RepoMetadataSidebarBlockSkeleton as ModuleMetadataSidebarBlockSkeleton };
