import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleSubmoduleDataQuery } from "../query";
import { ModuleOutputs as ModuleOutputsComponent } from "@/components/ModuleOutputs";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { ModuleSubmoduleMetaTags } from "../components/MetaTags";

export function ModuleSubmoduleOutputs() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  return (
    <>
      <ModuleSubmoduleMetaTags page="Outputs" />
      <ModuleOutputsComponent outputs={data.outputs} />
    </>
  );
}
