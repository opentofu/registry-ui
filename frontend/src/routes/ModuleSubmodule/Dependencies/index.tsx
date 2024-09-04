import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleSubmoduleDataQuery } from "../query";
import { ModuleDependencies as ModuleDependenciesComponent } from "@/components/ModuleDependencies";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { ModuleSubmoduleMetaTitle } from "../components/MetaTitle";

export function ModuleSubmoduleDependencies() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  return (
    <>
      <ModuleSubmoduleMetaTitle page="Dependencies" />
      <ModuleDependenciesComponent
        moduleDependencies={data.dependencies}
        providerDependencies={data.providers}
      />
    </>
  );
}
