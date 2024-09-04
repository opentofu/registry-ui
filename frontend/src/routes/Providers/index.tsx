import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense } from "react";

import { ProvidersList, ProvidersListSkeleton } from "./components/List";
import { MetaTitle } from "@/components/MetaTitle";

export function Providers() {
  return (
    <SimpleLayout>
      <MetaTitle>Providers</MetaTitle>
      <div className="mb-5 flex justify-between">
        <div className="flex flex-col gap-5">
          <PageTitle>Providers</PageTitle>
          <Paragraph>
            Providers are plugins to OpenTofu and create or destroy resources
            using their backing API based on your OpenTofu configuration.
          </Paragraph>
        </div>
        <Button
          target="_blank"
          rel="noopener noreferrer"
          variant="primary"
          href="https://github.com/opentofu/registry/issues/new?assignees=&labels=provider%2Csubmission&projects=&template=provider.yml&title=Provider%3A+"
        >
          Add provider
        </Button>
      </div>

      <Suspense fallback={<ProvidersListSkeleton />}>
        <ProvidersList />
      </Suspense>
    </SimpleLayout>
  );
}
