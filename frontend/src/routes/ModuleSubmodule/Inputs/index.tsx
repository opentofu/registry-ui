import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleSubmoduleDataQuery } from "../query";
import { ModuleInputs } from "@/components/ModuleInputs";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";

export function ModuleSubmoduleInputs() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  return <ModuleInputs inputs={data.variables} />;
}