import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleDataQuery } from "../query";

import { MetaTitle } from "@/components/MetaTitle";
import { useModuleParams } from "../hooks/useModuleParams";
import { Suspense } from "react";

interface ModuleMetaTitleProps {
  page?: string;
}

function ModuleMetaTitleContent({ page }: ModuleMetaTitleProps) {
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

  return <MetaTitle>{title}</MetaTitle>;
}

export function ModuleMetaTitle({ page }: ModuleMetaTitleProps) {
  return (
    <Suspense fallback={null}>
      <ModuleMetaTitleContent page={page} />
    </Suspense>
  );
}
