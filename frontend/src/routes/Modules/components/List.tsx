import { ModulesCardItem, ModulesCardItemSkeleton } from "./CardItem";

import { Virtuoso } from "react-virtuoso";
import { ComponentProps, forwardRef } from "react";

import { useSuspenseQuery } from "@tanstack/react-query";
import { getModulesQuery } from "../query";

const ModulesListWrapper = forwardRef<HTMLDivElement, ComponentProps<"div">>(
  function ModulesListWrapperImpl({ children, ...props }, ref) {
    return (
      <div ref={ref} className="flex flex-col gap-3" {...props}>
        {children}
      </div>
    );
  },
);

export function ModulesList() {
  const { data: modules } = useSuspenseQuery(getModulesQuery());

  return (
    <>
      <Virtuoso
        useWindowScroll
        data={modules}
        totalCount={modules.length}
        components={{
          List: ModulesListWrapper,
        }}
        itemContent={(_, module) => (
          <ModulesCardItem
            key={module.addr.namespace + module.addr.name + module.addr.target}
            addr={module.addr}
            description={module.description}
            latestVersion={module.versions[0]}
          />
        )}
      />
    </>
  );
}

export function ModulesListSkeleton() {
  return (
    <ModulesListWrapper>
      <ModulesCardItemSkeleton />
      <ModulesCardItemSkeleton />
      <ModulesCardItemSkeleton />
      <ModulesCardItemSkeleton />
      <ModulesCardItemSkeleton />
    </ModulesListWrapper>
  );
}
