import { defer } from "react-router-dom";
import { queryClient } from "@/query";
import { getModulesQuery } from "./query";

export const modulesLoader = () => {
  const index = queryClient.ensureQueryData(getModulesQuery());
  return defer({ index });
};
