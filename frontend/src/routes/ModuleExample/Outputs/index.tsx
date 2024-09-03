import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleExampleDataQuery } from "../query";
import { ModuleOutputs as ModuleOutputsComponent } from "@/components/ModuleOutputs";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { ModuleExampleMetaTags } from "../components/MetaTags";

export function ModuleExampleOutputs() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleExampleDataQuery(namespace, name, target, version, example),
  );

  return (
    <>
      <ModuleExampleMetaTags page="Outputs" />
      <ModuleOutputsComponent outputs={data.outputs} />
    </>
  );
}
