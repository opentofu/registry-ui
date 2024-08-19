import { ModulesCardItem } from "./CardItem";

import { Virtuoso } from "react-virtuoso";
import { forwardRef } from "react";

import { useSuspenseQuery } from "@tanstack/react-query";
import { getModulesQuery } from "../query";

export function ModulesList() {
  const { data: modules } = useSuspenseQuery(getModulesQuery());

  return (
    <>
      <Virtuoso
        className="z-0 grow"
        useWindowScroll
        data={modules}
        totalCount={modules.length}
        components={{
          List: forwardRef(({ children, ...props }, ref) => (
            <div ref={ref} className="flex flex-col gap-3" {...props}>
              {children}
            </div>
          )),
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
