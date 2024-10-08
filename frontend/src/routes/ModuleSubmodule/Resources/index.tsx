import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleSubmoduleDataQuery } from "../query";

import { ModuleResources as ModuleResourcesComponent } from "@/components/ModuleResources";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { ModuleSubmoduleMetaTags } from "../components/MetaTags";

export function ModuleSubmoduleResources() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  return (
    <>
      <ModuleSubmoduleMetaTags page="Resources" />
      <ModuleResourcesComponent resources={data.resources} />
    </>
  );
}
