import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";
import { ModuleInputs as ModuleInputsComponent } from "@/components/ModuleInputs";

export function ModuleInputs() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return <ModuleInputsComponent inputs={data.variables} />;
}
