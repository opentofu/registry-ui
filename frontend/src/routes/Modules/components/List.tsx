import { ModulesCardItem, ModulesCardItemSkeleton } from "./CardItem";

import { ComponentProps, forwardRef } from "react";
import { VirtuosoGrid } from "react-virtuoso";

import { useSuspenseQuery } from "@tanstack/react-query";
import { getModulesQuery } from "../query";

const gridComponents = {
  List: forwardRef<HTMLDivElement, ComponentProps<"div">>(({ style, children, ...props }, ref) => (
    <div
      ref={ref}
      {...props}
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        ...style,
      }}
    >
      {children}
    </div>
  )),
  Item: ({ children, ...props }: ComponentProps<"div">) => (
    <div
      {...props}
      className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2 flex"
      style={{
        display: 'flex',
        flex: 'none',
        boxSizing: 'border-box',
      }}
    >
      <div className="flex w-full">
        {children}
      </div>
    </div>
  ),
};

interface ModulesListProps {
  searchQuery?: string;
}

export function ModulesList({ searchQuery = "" }: ModulesListProps) {
  const { data: modules } = useSuspenseQuery(getModulesQuery());

  const filteredModules = modules.filter((module) => {
    // First filter out modules without versions
    if (!module.versions || module.versions.length === 0) return false;
    
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    const fullName = `${module.addr.namespace}/${module.addr.name}/${module.addr.target}`.toLowerCase();
    const displayName = module.addr.display?.toLowerCase() || fullName;
    
    return (
      fullName.includes(query) ||
      displayName.includes(query) ||
      module.addr.namespace.toLowerCase().includes(query) ||
      module.addr.name.toLowerCase().includes(query) ||
      module.addr.target.toLowerCase().includes(query) ||
      module.description.toLowerCase().includes(query)
    );
  });

  if (filteredModules.length === 0) {
    return (
      <div className="text-center py-12">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">No modules found</h3>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Try adjusting your search terms or <a href="https://github.com/opentofu/registry/issues/new?assignees=&labels=module%2Csubmission&projects=&template=module.yml&title=Module%3A+" className="text-brand-500 hover:text-brand-600">add a new module</a>.
        </p>
      </div>
    );
  }

  return (
    <VirtuosoGrid
      useWindowScroll
      data={filteredModules}
      totalCount={filteredModules.length}
      components={gridComponents}
      itemContent={(index) => {
        const module = filteredModules[index];
        return <ModulesCardItem module={module} />;
      }}
    />
  );
}

export function ModulesListSkeleton() {
  return (
    <div className="flex flex-wrap">
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
      <div className="w-full md:w-1/2 lg:w-1/3 xl:w-1/4 p-2">
        <ModulesCardItemSkeleton />
      </div>
    </div>
  );
}