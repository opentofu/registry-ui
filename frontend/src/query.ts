import { QueryClient } from "@tanstack/react-query";
import ky from "ky";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 10,
    },
  },
});

export const api = ky.create({ prefixUrl: import.meta.env.VITE_DATA_API_URL });
