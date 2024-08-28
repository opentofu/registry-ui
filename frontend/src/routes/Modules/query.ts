import { queryOptions } from "@tanstack/react-query";
import { definitions } from "@/api";
import { api } from "@/query";

export const getModulesQuery = () =>
  queryOptions({
    queryKey: ["modules"],
    queryFn: async () => {
      const data =
        await api(`modules/index.json`).json<definitions["ModuleList"]>();

      return data.modules;
    },
  });
