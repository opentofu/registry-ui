import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleDataQuery } from "../query";

import { MetaTags } from "@/components/MetaTags";
import { useModuleParams } from "../hooks/useModuleParams";
import { Suspense } from "react";

interface ModuleMetaTagsProps {
  page?: string;
}

function ModuleMetaTagsContent({ page }: ModuleMetaTagsProps) {
  const { namespace, name, target, version, isLatest } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  let title = `${data.addr.namespace}/${data.addr.name}/${data.addr.target}`;

  if (!isLatest) {
    title = `${version} - ${title}`;
  }

  if (page) {
    title = `${page} - ${title}`;
  }

  return <MetaTags title={title} description={data.description} />;
}

export function ModuleMetaTags({ page }: ModuleMetaTagsProps) {
  return (
    <Suspense fallback={null}>
      <ModuleMetaTagsContent page={page} />
    </Suspense>
  );
}
