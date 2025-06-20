import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { LicenseSidebarBlock } from "@/components/LicenseSidebarBlock";

export function ProviderMetadataSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  return <LicenseSidebarBlock license={data.license} />;
}

export function ProviderMetadataSidebarBlockSkeleton() {
  return <LicenseSidebarBlock />;
}
