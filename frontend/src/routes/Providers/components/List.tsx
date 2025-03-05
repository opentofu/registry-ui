import { ProvidersCardItem, ProvidersCardItemSkeleton } from "./CardItem";

import { Virtuoso } from "react-virtuoso";
import { ComponentProps, forwardRef } from "react";

import { useSuspenseQuery } from "@tanstack/react-query";
import { getProvidersQuery } from "../query";

const ProvidersListWrapper = forwardRef<HTMLDivElement, ComponentProps<"div">>(
  function ProvidersListWrapperImpl({ children, ...props }, ref) {
    return (
      <div ref={ref} className="flex flex-col gap-3" {...props}>
        {children}
      </div>
    );
  },
);

export function ProvidersList() {
  const { data: providers } = useSuspenseQuery(getProvidersQuery());

  return (
    <Virtuoso
      useWindowScroll
      data={providers}
      totalCount={providers.length}
      components={{
        List: ProvidersListWrapper,
      }}
      itemContent={(_, provider) => (
        provider.versions && (
          <ProvidersCardItem
          key={provider.addr.namespace + provider.addr.name}
          addr={provider.addr}
          description={provider.description}
          latestVersion={provider.versions[0]}
          />
        )
      )}
    />
  );
}

export function ProvidersListSkeleton() {
  return (
    <ProvidersListWrapper>
      <ProvidersCardItemSkeleton />
      <ProvidersCardItemSkeleton />
      <ProvidersCardItemSkeleton />
      <ProvidersCardItemSkeleton />
      <ProvidersCardItemSkeleton />
    </ProvidersListWrapper>
  );
}
