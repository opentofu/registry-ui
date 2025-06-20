import { queryClient } from "@/query";
import { getModulesQuery } from "./query";

export const modulesLoader = () => {
  const index = queryClient.ensureQueryData(getModulesQuery());
  return { index };
};
