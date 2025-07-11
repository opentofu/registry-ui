import { LoaderFunction } from "react-router";

import { ModuleExampleRouteContext } from "./types";

export const moduleExampleMiddleware: LoaderFunction = async (
  args,
  context,
) => {
  const moduleContext = context as ModuleExampleRouteContext;

  moduleContext.example = args.params.example;
};
