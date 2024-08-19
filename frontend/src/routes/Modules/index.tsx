import { Suspense } from "react";
import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";
import { ModulesCardItemSkeleton } from "./components/CardItem";
import { ModulesList } from "./components/List";

export function Modules() {
  return (
    <SimpleLayout>
      <div className="mb-5 flex justify-between">
        <div className="flex flex-col gap-5">
          <PageTitle>Modules</PageTitle>
          <Paragraph>
              Modules are reusable packages of OpenTofu code to speed up development.
          </Paragraph>
        </div>
        <Button variant="primary" href="https://opentofu.org">
          Add module
        </Button>
      </div>
      <Suspense
        fallback={
          <div className="flex flex-col gap-3">
            <ModulesCardItemSkeleton />
            <ModulesCardItemSkeleton />
            <ModulesCardItemSkeleton />
            <ModulesCardItemSkeleton />
            <ModulesCardItemSkeleton />
          </div>
        }
      >
        <ModulesList />
      </Suspense>
    </SimpleLayout>
  );
}
