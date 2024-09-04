import { useSuspenseQuery } from "@tanstack/react-query";

import { MetaTitle } from "@/components/MetaTitle";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { getModuleDataQuery } from "@/routes/Module/query";
import { Suspense } from "react";

interface ModuleMetaTitleProps {
  page?: string;
}

export function ModuleSubmoduleMetaTitleContent({
  page,
}: ModuleMetaTitleProps) {
  const { namespace, name, target, version, isLatest, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  let title = `Submodule: ${submodule} - ${data.addr.namespace}/${data.addr.name}/${data.addr.target}`;

  if (!isLatest) {
    title = `${version} - ${title}`;
  }

  if (page) {
    title = `${page} - ${title}`;
  }

  return <MetaTitle>{title}</MetaTitle>;
}

export function ModuleSubmoduleMetaTitle({ page }: ModuleMetaTitleProps) {
  return (
    <Suspense fallback={null}>
      <ModuleSubmoduleMetaTitleContent page={page} />
    </Suspense>
  );
}
