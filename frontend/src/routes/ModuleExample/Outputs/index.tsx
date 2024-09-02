import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleExampleDataQuery } from "../query";
import { ModuleOutputs as ModuleOutputsComponent } from "@/components/ModuleOutputs";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { ModuleExampleMetaTitle } from "../components/MetaTitle";

export function ModuleExampleOutputs() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleExampleDataQuery(namespace, name, target, version, example),
  );

  return (
    <>
      <ModuleExampleMetaTitle page="Outputs" />
      <ModuleOutputsComponent outputs={data.outputs} />
    </>
  );
}
