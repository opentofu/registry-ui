import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleSubmoduleDataQuery } from "../query";
import { ModuleInputs } from "@/components/ModuleInputs";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { ModuleSubmoduleMetaTitle } from "../components/MetaTitle";

export function ModuleSubmoduleInputs() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
  );

  return (
    <>
      <ModuleSubmoduleMetaTitle page="Inputs" />
      <ModuleInputs inputs={data.variables} />
    </>
  );
}
