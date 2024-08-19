import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleExampleDataQuery } from "../query";
import { ModuleInputs } from "@/components/ModuleInputs";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";

export function ModuleExampleInputs() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleExampleDataQuery(namespace, name, target, version, example),
  );

  return <ModuleInputs inputs={data.variables} />;
}
