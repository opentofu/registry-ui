import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleSubmoduleDataQuery } from "../query";
import { ModuleDependencies as ModuleDependenciesComponent } from "@/components/ModuleDependencies";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { ModuleSubmoduleMetaTags } from "../components/MetaTags";

export function ModuleSubmoduleDependencies() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  return (
    <>
      <ModuleSubmoduleMetaTags page="Dependencies" />
      <ModuleDependenciesComponent
        moduleDependencies={data.dependencies}
        providerDependencies={data.providers}
      />
    </>
  );
}
