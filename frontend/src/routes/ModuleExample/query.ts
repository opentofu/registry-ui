import { queryClient } from "@/query";
import { queryOptions } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../Module/query";
import { NotFoundPageError } from "@/utils/errors";

export const getModuleExampleReadmeQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
  example: string | undefined,
) => {
  return queryOptions({
    queryKey: [
      "module-example-readme",
      namespace,
      name,
      target,
      version,
      example,
    ],
    queryFn: async () => {
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/${namespace}/${name}/${target}/${version}/examples/${example}/README.md`,
      );

      return response.text();
    },
  });
};

export const getModuleExampleDataQuery = (
  namespace: string | undefined,
  name: string | undefined,
  target: string | undefined,
  version: string | undefined,
  example: string | undefined,
) => {
  return queryOptions({
    queryKey: ["module-example", namespace, name, target, version, example],
    queryFn: async () => {
      const data = await queryClient.ensureQueryData(
        getModuleVersionDataQuery(namespace, name, target, version),
      );

      if (!example || !data.examples[example]) {
        throw new NotFoundPageError();
      }

      return data.examples[example];
    },
  });
};
