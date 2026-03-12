import { useSuspenseQuery } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import {
  getModuleSubmoduleDataQuery,
  getModuleSubmoduleReadmeQuery,
} from "../query";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { Suspense } from "react";
import { ModuleSubmoduleMetaTags } from "../components/MetaTags";
import { EditLink } from "@/components/EditLink";
import { EmptyState } from "@/components/EmptyState";

function ModuleSubmoduleReadmeContentMarkdown({
  editLink,
}: {
  editLink: string | undefined;
}) {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleReadmeQuery(namespace, name, target, version, submodule),
  );

  return (
    <>
      <Markdown text={data} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

function ModuleSubmoduleReadmeContent() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data: submoduleData } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  if (!submoduleData.readme) {
    return (
      <EmptyState
        text="This submodule does not have a README."
        className="mt-5"
      />
    );
  }

  return (
    <ModuleSubmoduleReadmeContentMarkdown editLink={submoduleData.edit_link} />
  );
}

function ModuleSubmoduleReadmeContentSkeleton() {
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
