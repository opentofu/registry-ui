import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";
import { ModuleDependencies as ModuleDependenciesComponent } from "@/components/ModuleDependencies";
import { ModuleMetaTags } from "../components/MetaTags";

export function ModuleDependencies() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <>
      <ModuleMetaTags page="Dependencies" />
      <ModuleDependenciesComponent
        moduleDependencies={data.dependencies}
        providerDependencies={data.providers}
      />
    </>
  );
}
