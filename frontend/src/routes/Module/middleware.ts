import { queryClient } from "@/query";
import { LoaderFunction, redirect } from "react-router-dom";
import { getModuleDataQuery } from "./query";
import { ModuleRouteContext } from "./types";

export const moduleMiddleware: LoaderFunction = async ({ params }, context) => {
  const { namespace, name, target, version } = params;

  const data = await queryClient.ensureQueryData(
    getModuleDataQuery(namespace, name, target),
  );

  const [latestVersion] = data.versions;

  if (version === latestVersion.id || !version) {
    return redirect(`/module/${namespace}/${name}/${target}/latest`);
  }

  const moduleContext = context as ModuleRouteContext;

  moduleContext.version = version === "latest" ? latestVersion.id : version;
  moduleContext.rawVersion = version;
  moduleContext.namespace = namespace;
  moduleContext.name = name;
  moduleContext.target = target;
};
