import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { UnifiedLayout } from "@/components/UnifiedLayout";
import { SidebarPanel } from "@/components/SidebarPanel";

import { Suspense, useState } from "react";

import { ProvidersList, ProvidersListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { SearchInput } from "@/components/SearchInput";
import { ListingFilterSidebar } from "@/components/ListingFilterSidebar";
import { ListingStatsSidebar } from "@/components/ListingStatsSidebar";
import { useListingFilters } from "@/hooks/useListingFilters";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getProvidersQuery } from "./query";

const title = "Providers";

const description =
  "Providers are plugins to OpenTofu and create or destroy resources using their backing API based on your OpenTofu configuration.";

export function Providers() {
  const { data: providers } = useSuspenseQuery(getProvidersQuery());
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebouncedValue(searchQuery, 300);

  const {
    filteredItems: filteredProviders,
    namespaceOptions,
    stats,
    sortBy,
    setSortBy,
    selectedNamespaces,
    setSelectedNamespaces,
    showOnlyOfficial,
    setShowOnlyOfficial,
    hideForks,
    setHideForks,
  } = useListingFilters({
    items: providers,
    searchQuery: debouncedSearchQuery,
    getNamespace: (provider) => provider.addr.namespace,
    getName: (provider) => provider.addr.name,
    getDescription: (provider) => provider.description,
    getPopularity: (provider) => provider.popularity,
    getLastUpdate: (provider) => provider.versions?.[0]?.published || "",
    getForkOf: (provider) => provider.fork_of,
  });

  return (
    <UnifiedLayout
      sidebar={
        <SidebarPanel>
          <ListingFilterSidebar
            sortBy={sortBy}
            onSortChange={setSortBy}
            namespaces={namespaceOptions}
            selectedNamespaces={selectedNamespaces}
            onNamespaceChange={setSelectedNamespaces}
            showOnlyOfficial={showOnlyOfficial}
            onShowOnlyOfficialChange={setShowOnlyOfficial}
            hideForks={hideForks}
            onHideForksChange={setHideForks}
          />
        </SidebarPanel>
      }
      afterSidebar={
        <SidebarPanel>
          <ListingStatsSidebar
            type="providers"
            stats={stats}
            addButtonHref="https://github.com/opentofu/registry/issues/new?assignees=&labels=provider%2Csubmission&projects=&template=provider.yml&title=Provider%3A+"
          />
        </SidebarPanel>
      }
      useIDEHeader={true}
    >
      <MetaTags title={title} description={description} />
      <div className="flex flex-col gap-5 px-5">
        <div>
          <PageTitle className="mb-3">{title}</PageTitle>
          <Paragraph className="text-gray-600 dark:text-gray-300">
            {description}
          </Paragraph>
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

      <div className="p-5">
        <Suspense fallback={<ProvidersListSkeleton />}>
          <ProvidersList
            providers={filteredProviders}
          />
        </Suspense>
      </div>
    </UnifiedLayout>
  );
}
