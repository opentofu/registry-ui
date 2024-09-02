import { definitions } from "@/api";
import { api } from "@/query";
import { queryOptions } from "@tanstack/react-query";

export const getModuleVersionDataQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
) => {
  return queryOptions({
    queryKey: ["module-version", namespace, name, target, version],
    queryFn: async () => {
      const data = await api(
        `modules/${namespace}/${name}/${target}/${version}/index.json`,
      ).json<definitions["ModuleVersion"]>();

      return data;
    },
  });
};

export const getModuleDataQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
) => {
  return queryOptions({
    queryKey: ["module", namespace, name, target],
    queryFn: async () => {
      const data = await api(
        `modules/${namespace}/${name}/${target}/index.json`,
      ).json<definitions["Module"]>();

      return data;
    },
  });
};

export const getModuleReadmeQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
) => {
  return queryOptions({
    queryKey: ["module-readme", namespace, name, target, version],
    queryFn: async () => {
      const data = await api(
        `modules/${namespace}/${name}/${target}/${version}/README.md`,
      ).text();

      return data;
    },
  });
};
