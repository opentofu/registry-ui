import { useModuleParams } from "@/routes/Module/hooks/useModuleParams";
import { useParams } from "react-router";

export function useModuleSubmoduleParams() {
  const moduleParams = useModuleParams();

  const { submodule } = useParams<{
    submodule: string;
  }>();

  return {
    ...moduleParams,
    submodule,
  };
}
