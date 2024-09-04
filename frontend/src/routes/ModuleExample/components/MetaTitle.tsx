import { useSuspenseQuery } from "@tanstack/react-query";

import { MetaTitle } from "@/components/MetaTitle";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { getModuleDataQuery } from "@/routes/Module/query";
import { Suspense } from "react";

interface ModuleMetaTitleProps {
  page?: string;
}

function ModuleExampleMetaTitleContent({ page }: ModuleMetaTitleProps) {
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

  return <MetaTitle>{title}</MetaTitle>;
}

export function ModuleExampleMetaTitle({ page }: ModuleMetaTitleProps) {
  return (
    <Suspense fallback={null}>
      <ModuleExampleMetaTitleContent page={page} />
    </Suspense>
  );
}
