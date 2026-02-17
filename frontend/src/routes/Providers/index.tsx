import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense, useState } from "react";

import { ProvidersList, ProvidersListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { SearchInput } from "@/components/SearchInput";

const title = "Providers";

const description =
  "Providers are plugins to OpenTofu and create or destroy resources using their backing API based on your OpenTofu configuration.";

export function Providers() {
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebouncedValue(searchQuery, 300);

  return (
    <SimpleLayout>
      <MetaTags title={title} description={description} />
      <div className="mb-8">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <PageTitle className="mb-3">{title}</PageTitle>
            <Paragraph className="text-gray-600 dark:text-gray-300">
              {description}
            </Paragraph>
          </div>
          <Button
            target="_blank"
            rel="noopener noreferrer"
            variant="primary"
            href="https://github.com/opentofu/registry/issues/new?assignees=&labels=provider%2Csubmission&projects=&template=provider.yml&title=Provider%3A+"
            className="flex-shrink-0"
          >
            Add provider
          </Button>
        </div>

        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search providers..."
          size="large"
          onKeyDown={(e) => {
            // Prevent the global search from capturing the "/" key
            if (e.key === "/") {
              e.stopPropagation();
            }
          }}
        />
      </div>

      <Suspense fallback={<ProvidersListSkeleton />}>
        <ProvidersList searchQuery={debouncedSearchQuery} />
      </Suspense>
    </SimpleLayout>
  );
}
