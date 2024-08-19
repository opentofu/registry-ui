import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense } from "react";

import { ProvidersList } from "./components/List";
import { ProvidersCardItemSkeleton } from "./components/CardItem";

export function Providers() {
  return (
    <SimpleLayout>
      <div className="mb-5 flex justify-between">
        <div className="flex flex-col gap-5">
          <PageTitle>Providers</PageTitle>
          <Paragraph>
              Providers are plugins to OpenTofu and create or destroy resources using their backing API based on your OpenTofu configuration.
          </Paragraph>
        </div>
        <Button variant="primary" href="https://opentofu.org">
          Add provider
        </Button>
      </div>

      <Suspense
        fallback={
          <div className="flex flex-col gap-3">
            <ProvidersCardItemSkeleton />
            <ProvidersCardItemSkeleton />
            <ProvidersCardItemSkeleton />
            <ProvidersCardItemSkeleton />
            <ProvidersCardItemSkeleton />
          </div>
        }
      >
        <ProvidersList />
      </Suspense>
    </SimpleLayout>
  );
}
