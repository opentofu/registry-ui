import { ProvidersCardItem } from "./CardItem";

import { Virtuoso } from "react-virtuoso";
import { forwardRef } from "react";

import { useSuspenseQuery } from "@tanstack/react-query";
import { getProvidersQuery } from "../query";

export function ProvidersList() {
  const { data: providers } = useSuspenseQuery(getProvidersQuery());

  return (
    <>
      <Virtuoso
        className="z-0 grow"
        useWindowScroll
        data={providers}
        totalCount={providers.length}
        components={{
          List: forwardRef(({ children, ...props }, ref) => (
            <div ref={ref} className="flex flex-col gap-3" {...props}>
              {children}
            </div>
          )),
        }}
        itemContent={(_, provider) => (
          <ProvidersCardItem
            key={provider.addr.namespace + provider.addr.name}
            addr={provider.addr}
            description={provider.description}
            latestVersion={provider.versions[0]}
          />
        )}
      />
    </>
  );
}
