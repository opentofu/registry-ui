import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleExampleDataQuery } from "../query";
import { ModuleInputs } from "@/components/ModuleInputs";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { ModuleExampleMetaTags } from "../components/MetaTags";

export function ModuleExampleInputs() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleExampleDataQuery(namespace, name, target, version, example),
  );

  return (
    <>
      <ModuleExampleMetaTags page="Inputs" />
      <ModuleInputs inputs={data.variables} />
    </>
  );
}
