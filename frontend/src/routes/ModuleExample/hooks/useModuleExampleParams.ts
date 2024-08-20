import { useModuleParams } from "@/routes/Module/hooks/useModuleParams";
import { useParams } from "react-router-dom";

export function useModuleExampleParams() {
  const moduleParams = useModuleParams();

  const { example } = useParams<{
    example: string;
  }>();

  return {
    ...moduleParams,
    example,
  };
}
