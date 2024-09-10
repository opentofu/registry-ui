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
