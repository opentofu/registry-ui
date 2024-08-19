import { queryOptions } from "@tanstack/react-query";

export const getModuleExampleReadmeQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
  example: string | undefined,
) => {
  return queryOptions({
    queryKey: ["module-readme", namespace, name, target, version, example],
    queryFn: async () => {
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/${namespace}/${name}/${target}/${version}/examples/${example}/README.md`,
      );

      return response.text();
    },
  });
};
