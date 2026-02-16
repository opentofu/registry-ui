import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { UnifiedLayout } from "@/components/UnifiedLayout";
import { SidebarPanel } from "@/components/SidebarPanel";

import { Suspense, useState } from "react";

import { ModulesList, ModulesListSkeleton } from "./components/List";
import { MetaTags } from "@/components/MetaTags";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { SearchInput } from "@/components/SearchInput";
import { ListingFilterSidebar } from "@/components/ListingFilterSidebar";
import { ListingStatsSidebar } from "@/components/ListingStatsSidebar";
import { useListingFilters } from "@/hooks/useListingFilters";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getModulesQuery } from "./query";

const title = "Modules";

const description =
	"Modules are reusable packages of OpenTofu code to speed up development.";

export function Modules() {
	const { data: modules } = useSuspenseQuery(getModulesQuery());
	const [searchQuery, setSearchQuery] = useState("");
	const debouncedSearchQuery = useDebouncedValue(searchQuery, 300);

	const {
		filteredItems: filteredModules,
		stats,
		sortBy,
		setSortBy,
	} = useListingFilters({
		items: modules,
		searchQuery: debouncedSearchQuery,
		getNamespace: (module) => module.addr.namespace,
		getName: (module) => module.addr.name,
		getDescription: (module) => module.description,
		getPopularity: (module) => module.popularity,
		getLastUpdate: (module) => module.versions?.[0]?.published || "",
		getForkOf: (module) => module.fork_of,
		getDisplayName: (module) => module.addr.display,
	});

	return (
		<UnifiedLayout
			sidebar={
				<SidebarPanel>
					<ListingFilterSidebar sortBy={sortBy} onSortChange={setSortBy} />
				</SidebarPanel>
			}
			afterSidebar={
				<SidebarPanel>
					<ListingStatsSidebar
						type="modules"
						stats={stats}
						addButtonHref="https://github.com/opentofu/registry/issues/new?assignees=&labels=module%2Csubmission&projects=&template=module.yml&title=Module%3A+"
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

			<div className="p-5">
				<Suspense fallback={<ModulesListSkeleton />}>
					<ModulesList modules={filteredModules} />
				</Suspense>
			</div>
		</UnifiedLayout>
	);
}
