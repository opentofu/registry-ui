import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense, useState } from "react";

import { ProvidersList, ProvidersListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";

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

        <div className="relative">
          <input
            type="text"
            aria-label="Query"
            placeholder="Search providers..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => {
              // Prevent the global search from capturing the "/" key
              if (e.key === "/") {
                e.stopPropagation();
              }
            }}
            className="focus:ring-brand-500 w-full rounded-lg border border-gray-200 bg-white px-4 py-3 pr-4 pl-12 text-sm focus:border-transparent focus:ring-2 focus:outline-none dark:border-gray-700 dark:bg-blue-900 dark:text-gray-200 dark:placeholder-gray-400"
          />
          <svg
            className="absolute top-1/2 left-4 h-5 w-5 -translate-y-1/2 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </div>
      </div>

      <Suspense fallback={<ProvidersListSkeleton />}>
        <ProvidersList searchQuery={debouncedSearchQuery} />
      </Suspense>
    </SimpleLayout>
  );
}
