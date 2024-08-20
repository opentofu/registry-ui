import { queryClient } from "@/query";
import { queryOptions } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../Module/query";
import { NotFoundPageError } from "@/utils/errors";

export const getModuleSubmoduleReadmeQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
  submodule: string | undefined,
) => {
  return queryOptions({
    queryKey: [
      "module-submodule-readme",
      namespace,
      name,
      target,
      version,
      submodule,
    ],
    queryFn: async () => {
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/${namespace}/${name}/${target}/${version}/modules/${submodule}/README.md`,
      );

      return response.text();
    },
  });
};

export const getModuleSubmoduleDataQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
  submodule: string | undefined,
) => {
  return queryOptions({
    queryKey: ["module-submodule", namespace, name, target, version, submodule],
    queryFn: async () => {
      const data = await queryClient.ensureQueryData(
        getModuleVersionDataQuery(namespace, name, target, version),
      );

      if (!submodule || !data.submodules[submodule]) {
        throw new NotFoundPageError();
      }

      return data.submodules[submodule];
    },
  });
};
