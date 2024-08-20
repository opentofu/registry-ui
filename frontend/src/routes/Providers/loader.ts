import { defer } from "react-router-dom";
import { queryClient } from "../../query";
import { getProvidersQuery } from "./query";

export const providersLoader = () => {
  const index = queryClient.ensureQueryData(getProvidersQuery());
  return defer({ index });
};
