import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "./query";
import "./index.css";
import { router } from "./router";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} future={{ v7_startTransition: true }} />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  </StrictMode>,
);
