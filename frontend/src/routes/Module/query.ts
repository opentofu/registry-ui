import { definitions } from "@/api";
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
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/${namespace}/${name}/${target}/${version}/index.json`,
      );

      const data = await response.json();

      return data as definitions["ModuleVersion"];
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
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/${namespace}/${name}/${target}/index.json`,
      );

      const data = await response.json();

      return data as definitions["Module"];
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
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/${namespace}/${name}/${target}/${version}/README.md`,
      );

      return response.text();
    },
  });
};
