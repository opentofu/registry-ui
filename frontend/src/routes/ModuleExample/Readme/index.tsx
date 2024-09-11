import { useSuspenseQueries, useSuspenseQuery } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import {
  getModuleExampleDataQuery,
  getModuleExampleReadmeQuery,
} from "../query";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { Suspense } from "react";
import { ModuleExampleMetaTags } from "../components/MetaTags";
import { EditLink } from "@/components/EditLink";

function ModuleExampleReadmeContent() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const [{ data }, { data: exampleData }] = useSuspenseQueries({
    queries: [
      getModuleExampleReadmeQuery(namespace, name, target, version, example),
      getModuleExampleDataQuery(namespace, name, target, version, example),
    ],
  });

  const editLink = exampleData.edit_link;

  return (
    <>
      <Markdown text={data} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

function ModuleExampleReadmeContentSkeleton() {
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

export function ModuleExampleReadme() {
  return (
    <div className="p-5">
      <ModuleExampleMetaTags />
      <Suspense fallback={<ModuleExampleReadmeContentSkeleton />}>
        <ModuleExampleReadmeContent />
      </Suspense>
    </div>
  );
}
