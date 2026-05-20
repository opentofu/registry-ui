import { queryOptions } from "@tanstack/react-query";
import { components } from "@/api";
import { api } from "@/query";

export const getModulesQuery = () =>
  queryOptions({
    queryKey: ["modules"],
    queryFn: async () => {
      const data =
        await api(`modules/index.json`).json<
          components["schemas"]["ModuleList"]
        >();

      return data.modules;
    },
  });
