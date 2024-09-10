import { useSuspenseQueries } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import {
  getModuleSubmoduleDataQuery,
  getModuleSubmoduleReadmeQuery,
} from "../query";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { Suspense } from "react";
import { ModuleSubmoduleMetaTags } from "../components/MetaTags";
import { EditLink } from "@/components/EditLink";

function ModuleSubmoduleReadmeContent() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const [{ data }, { data: submoduleData }] = useSuspenseQueries({
    queries: [
      getModuleSubmoduleReadmeQuery(
        namespace,
        name,
        target,
        version,
        submodule,
      ),
      getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
    ],
  });

  const editLink = submoduleData.edit_link;

  return (
    <>
      <Markdown text={data} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

function ModuleSubmoduleReadmeContentSkeleton() {
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

export function ModuleSubmoduleReadme() {
  return (
    <div className="p-5">
      <ModuleSubmoduleMetaTags />
      <Suspense fallback={<ModuleSubmoduleReadmeContentSkeleton />}>
        <ModuleSubmoduleReadmeContent />
      </Suspense>
    </div>
  );
}
