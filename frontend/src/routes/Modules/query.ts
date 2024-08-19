import { queryOptions } from "@tanstack/react-query";
import { definitions } from "@/api";

export const getModulesQuery = () =>
  queryOptions({
    queryKey: ["modules"],
    queryFn: async () => {
      const response = await fetch(
        `${import.meta.env.VITE_DATA_API_URL}/modules/index.json`,
      );

      const res = await response.json();

      return res.modules as definitions["ModuleList"]["modules"];
    },
  });
