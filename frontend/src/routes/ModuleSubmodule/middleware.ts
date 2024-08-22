import { LoaderFunction } from "react-router-dom";

import { ModuleSubmoduleRouteContext } from "./types";

export const moduleSubmoduleMiddleware: LoaderFunction = async (
  args,
  context,
) => {
  const moduleContext = context as ModuleSubmoduleRouteContext;

  moduleContext.submodule = args.params.submodule;
};
