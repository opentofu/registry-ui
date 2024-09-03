import { useSuspenseQuery } from "@tanstack/react-query";

import { MetaTags } from "@/components/MetaTags";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { getModuleDataQuery } from "@/routes/Module/query";
import { Suspense } from "react";

interface ModuleMetaTagsProps {
  page?: string;
}

function ModuleExampleMetaTagsContent({ page }: ModuleMetaTagsProps) {
  const { namespace, name, target, version, isLatest, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  let title = `Example: ${example} - ${data.addr.namespace}/${data.addr.name}/${data.addr.target}`;

  if (!isLatest) {
    title = `${version} - ${title}`;
  }

  if (page) {
    title = `${page} - ${title}`;
  }

  return <MetaTags title={title} description={data.description} />;
}

export function ModuleExampleMetaTags({ page }: ModuleMetaTagsProps) {
  return (
    <Suspense fallback={null}>
      <ModuleExampleMetaTagsContent page={page} />
    </Suspense>
  );
}
