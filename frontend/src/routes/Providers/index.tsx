import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense } from "react";

import { ProvidersList, ProvidersListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";

const title = "Providers";

const description =
  "Providers are plugins to OpenTofu and create or destroy resources using their backing API based on your OpenTofu configuration.";

export function Providers() {
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
