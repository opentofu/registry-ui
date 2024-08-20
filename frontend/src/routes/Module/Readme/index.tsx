import { useSuspenseQuery } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import { getModuleReadmeQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";
import { Suspense } from "react";

export function ModuleReadmeContent() {
  const { namespace, name, version, target } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleReadmeQuery(namespace, name, target, version),
  );

  return <Markdown text={data} />;
}

export function ModuleReadmeContentSkeleton() {
  return (
    <>
      <span className="mt-6 flex h-em w-52 animate-pulse bg-gray-500/25 text-4xl" />

      <span className="mt-5 flex h-em w-[500px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[400px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[450px] animate-pulse bg-gray-500/25" />

      <span className="mt-8 flex h-em w-[300px] animate-pulse bg-gray-500/25 text-3xl" />

      <span className="mt-5 flex h-em w-[600px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[550px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[350px] animate-pulse bg-gray-500/25" />
    </>
  );
}

export function ModuleReadme() {
  return (
    <div className="p-5">
      <Suspense fallback={<ModuleReadmeContentSkeleton />}>
        <ModuleReadmeContent />
      </Suspense>
    </div>
  );
}
