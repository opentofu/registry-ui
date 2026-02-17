import { Button } from "@/components/Button";
import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { SimpleLayout } from "@/components/SimpleLayout";

import { Suspense, useState } from "react";

import { ModulesList, ModulesListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { SearchInput } from "@/components/SearchInput";

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

        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search modules..."
          size="large"
          onKeyDown={(e) => {
            // Prevent the global search from capturing the "/" key
            if (e.key === "/") {
              e.stopPropagation();
            }
          }}
        />
      </div>

      <Suspense fallback={<ModulesListSkeleton />}>
        <ModulesList searchQuery={debouncedSearchQuery} />
      </Suspense>
    </SimpleLayout>
  );
}