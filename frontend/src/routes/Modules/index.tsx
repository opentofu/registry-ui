import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense, useState } from "react";

import { ModulesList, ModulesListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";

const title = "Modules";

const description =
  "Modules are reusable packages of OpenTofu code to speed up development.";

export function Modules() {
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebouncedValue(searchQuery, 300);

  return (
    <SimpleLayout>
      <MetaTags title={title} description={description} />
      <div className="mb-8">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <PageTitle className="mb-3">{title}</PageTitle>
            <Paragraph className="text-gray-600 dark:text-gray-300">{description}</Paragraph>
          </div>
          <Button
            target="_blank"
            rel="noopener noreferrer"
            variant="primary"
            href="https://github.com/opentofu/registry/issues/new?assignees=&labels=module%2Csubmission&projects=&template=module.yml&title=Module%3A+"
            className="flex-shrink-0"
          >
            Add module
          </Button>
        </div>

        <div className="relative">
          <input
            type="text"
            placeholder="Search modules..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => {
              // Prevent the global search from capturing the "/" key
              if (e.key === "/") {
                e.stopPropagation();
              }
            }}
            className="w-full px-4 py-3 pl-12 pr-4 text-sm bg-white border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent dark:bg-blue-900 dark:border-gray-700 dark:text-gray-200 dark:placeholder-gray-400"
          />
          <svg
            className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
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

      <Suspense fallback={<ModulesListSkeleton />}>
        <ModulesList searchQuery={debouncedSearchQuery} />
      </Suspense>
    </SimpleLayout>
  );
}