import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";
import { ModuleDependencies as ModuleDependenciesComponent } from "@/components/ModuleDependencies";
import { ModuleMetaTitle } from "../components/MetaTitle";

export function ModuleDependencies() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <>
      <ModuleMetaTitle page="Dependencies" />
      <ModuleDependenciesComponent
        moduleDependencies={data.dependencies}
        providerDependencies={data.providers}
      />
    </>
  );
}
