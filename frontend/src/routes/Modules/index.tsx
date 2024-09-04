import { Suspense } from "react";
import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";
import { ModulesList, ModulesListSkeleton } from "./components/List";
import { MetaTitle } from "@/components/MetaTitle";

export function Modules() {
  return (
    <SimpleLayout>
      <MetaTitle>Modules</MetaTitle>
      <div className="mb-5 flex justify-between">
        <div className="flex flex-col gap-5">
          <PageTitle>Modules</PageTitle>
          <Paragraph>
            Modules are reusable packages of OpenTofu code to speed up
            development.
          </Paragraph>
        </div>
        <Button
          target="_blank"
          rel="noopener noreferrer"
          variant="primary"
          href="https://github.com/opentofu/registry/issues/new?assignees=&labels=module%2Csubmission&projects=&template=module.yml&title=Module%3A+"
        >
          Add module
        </Button>
      </div>
      <Suspense fallback={<ModulesListSkeleton />}>
        <ModulesList />
      </Suspense>
    </SimpleLayout>
  );
}
