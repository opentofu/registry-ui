import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";
import { ModuleOutputs as ModuleOutputsComponent } from "@/components/ModuleOutputs";
import { ModuleMetaTags } from "../components/MetaTags";

export function ModuleOutputs() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <>
      <ModuleMetaTags page="Outputs" />
      <ModuleOutputsComponent outputs={data.outputs} />
    </>
  );
}
