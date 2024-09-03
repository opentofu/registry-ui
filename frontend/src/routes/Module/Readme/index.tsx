import { useSuspenseQueries } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import { getModuleReadmeQuery, getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";
import { Suspense } from "react";
import { EditLink } from "@/components/EditLink";
import { ModuleMetaTags } from "../components/MetaTags";

function ModuleReadmeContent() {
  const { namespace, name, version, target } = useModuleParams();

  const [{ data }, { data: versionData }] = useSuspenseQueries({
    queries: [
      getModuleReadmeQuery(namespace, name, target, version),
      getModuleVersionDataQuery(namespace, name, target, version),
    ],
  });

  const editLink = versionData.edit_link;

  return (
    <>
      <Markdown text={data} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

function ModuleReadmeContentSkeleton() {
  return (
    <>
      <span className="flex h-em w-52 animate-pulse bg-gray-500/25 text-4xl" />

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
      <ModuleMetaTags />
      <Suspense fallback={<ModuleReadmeContentSkeleton />}>
        <ModuleReadmeContent />
      </Suspense>
    </div>
  );
}
