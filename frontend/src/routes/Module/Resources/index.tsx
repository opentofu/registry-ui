import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";

import { useModuleParams } from "../hooks/useModuleParams";
import { ModuleResources as ModuleResourcesComponent } from "@/components/ModuleResources";
import { ModuleMetaTags } from "../components/MetaTags";

export function ModuleResources() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <>
      <ModuleMetaTags page="Resources" />
      <ModuleResourcesComponent resources={data.resources} />
    </>
  );
}
