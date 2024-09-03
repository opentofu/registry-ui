import { Suspense } from "react";
import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";
import { ModulesList, ModulesListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";

const title = "Modules";

const description =
  "Modules are reusable packages of OpenTofu code to speed up development.";

export function Modules() {
  return (
    <SimpleLayout>
      <MetaTags title={title} description={description} />
      <div className="mb-5 flex justify-between">
        <div className="flex flex-col gap-5">
          <PageTitle>{title}</PageTitle>
          <Paragraph>{description}</Paragraph>
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
