import { useSuspenseQuery } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import {
  getModuleExampleDataQuery,
  getModuleExampleReadmeQuery,
} from "../query";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { Suspense } from "react";
import { ModuleExampleMetaTags } from "../components/MetaTags";
import { EditLink } from "@/components/EditLink";
import { EmptyState } from "@/components/EmptyState";

function ModuleExampleReadmeContentMarkdown({
  editLink,
}: {
  editLink: string | undefined;
}) {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleExampleReadmeQuery(namespace, name, target, version, example),
  );

  return (
    <>
      <Markdown text={data} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

function ModuleExampleReadmeContent() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data: exampleData } = useSuspenseQuery(
    getModuleExampleDataQuery(namespace, name, target, version, example),
  );

  if (!exampleData.readme) {
    return (
      <EmptyState
        text="This example does not have a README."
        className="mt-5"
      />
    );
  }

  return (
    <ModuleExampleReadmeContentMarkdown editLink={exampleData.edit_link} />
  );
}

function ModuleExampleReadmeContentSkeleton() {
  return (
    <>
      <span className="h-em mt-6 flex w-52 animate-pulse bg-gray-500/25 text-4xl" />

      <span className="h-em mt-5 flex w-[500px] animate-pulse bg-gray-500/25" />
      <span className="h-em mt-1 flex w-[400px] animate-pulse bg-gray-500/25" />
      <span className="h-em mt-1 flex w-[450px] animate-pulse bg-gray-500/25" />

      <span className="h-em mt-8 flex w-[300px] animate-pulse bg-gray-500/25 text-3xl" />

      <span className="h-em mt-5 flex w-[600px] animate-pulse bg-gray-500/25" />
      <span className="h-em mt-1 flex w-[550px] animate-pulse bg-gray-500/25" />
      <span className="h-em mt-1 flex w-[350px] animate-pulse bg-gray-500/25" />
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
