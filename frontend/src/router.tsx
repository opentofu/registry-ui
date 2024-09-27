import { createBrowserRouter, Navigate } from "react-router-dom";
import { Home } from "./routes/Home";
import { Modules } from "./routes/Modules";
import { Providers } from "./routes/Providers";

import { createCrumb } from "./crumbs";
import { Error } from "./routes/Error";
import { Module } from "./routes/Module";
import { ModuleDependencies } from "./routes/Module/Dependencies";
import { ModuleInputs } from "./routes/Module/Inputs";
import { ModuleOutputs } from "./routes/Module/Outputs";
import { ModuleReadme } from "./routes/Module/Readme";
import { ModuleResources } from "./routes/Module/Resources";
import { ModuleExample } from "./routes/ModuleExample";
import { ModuleExampleInputs } from "./routes/ModuleExample/Inputs";
import { ModuleExampleOutputs } from "./routes/ModuleExample/Outputs";
import { ModuleExampleReadme } from "./routes/ModuleExample/Readme";
import { ModuleSubmodule } from "./routes/ModuleSubmodule";
import { ModuleSubmoduleDependencies } from "./routes/ModuleSubmodule/Dependencies";
import { ModuleSubmoduleInputs } from "./routes/ModuleSubmodule/Inputs";
import { ModuleSubmoduleOutputs } from "./routes/ModuleSubmodule/Outputs";
import { ModuleSubmoduleReadme } from "./routes/ModuleSubmodule/Readme";
import { ModuleSubmoduleResources } from "./routes/ModuleSubmodule/Resources";
import { Provider } from "./routes/Provider";
import { ProviderDocs } from "./routes/Provider/Docs";
import { ProviderOverview } from "./routes/Provider/Overview";
import { providersLoader } from "./routes/Providers/loader";
import { providerOverviewLoader } from "./routes/Provider/Overview/loader";
import { providerDocsLoader } from "./routes/Provider/Docs/loader";
import { providerLoader } from "./routes/Provider/loader";
import { modulesLoader } from "./routes/Modules/loader";
import { providerMiddleware } from "./routes/Provider/middleware";
import {
  moduleMetadataMiddleware,
  moduleMiddleware,
} from "./routes/Module/middleware";
import { moduleLoader } from "./routes/Module/loader";
import { moduleReadmeLoader } from "./routes/Module/Readme/loader";
import { moduleExampleLoader } from "./routes/ModuleExample/loader";
import { moduleExampleReadmeLoader } from "./routes/ModuleExample/Readme/loader";
import { ModuleRouteContext } from "./routes/Module/types";
import { moduleExampleMiddleware } from "./routes/ModuleExample/middleware";
import { ModuleExampleRouteContext } from "./routes/ModuleExample/types";
import { ProviderRouteContext } from "./routes/Provider/types";
import { moduleSubmoduleLoader } from "./routes/ModuleSubmodule/loader";
import { ModuleSubmoduleRouteContext } from "./routes/ModuleSubmodule/types";
import { moduleSubmoduleReadmeLoader } from "./routes/ModuleSubmodule/Readme/loader";
import { moduleSubmoduleMiddleware } from "./routes/ModuleSubmodule/middleware";
import { ProviderError } from "./routes/Provider/components/Error";
import { Docs } from "./routes/Docs";
import { docsLoader } from "./routes/Docs/loader";

export const router = createBrowserRouter(
  [
    {
      errorElement: <Error />,
      children: [
        {
          id: "home",
          index: true,
          element: <Home />,
        },
        {
          id: "docs",
          path: "/docs/*",
          element: <Docs />,
          loader: docsLoader,
        },
        {
          id: "providers",
          path: "/providers",
          element: <Providers />,
          loader: providersLoader,
          handle: {
            crumb: () => createCrumb("/providers", "Providers"),
          },
        },
        {
          id: "modules",
          path: "/modules",
          element: <Modules />,
          loader: modulesLoader,
          handle: {
            crumb: () => createCrumb("/modules", "Modules"),
          },
        },
        {
          id: "module",
          path: "module",
          handle: {
            crumb: () => createCrumb("/modules", "Modules"),
          },
          children: [
            {
              path: ":namespace",
              children: [
                {
                  index: true,
                  element: <Navigate to="/modules" />,
                },
                {
                  id: "module-version",
                  path: ":name/:target/:version?",
                  loader: moduleLoader,
                  handle: {
                    middleware: moduleMiddleware,
                    crumb: ({
                      namespace,
                      name,
                      target,
                      rawVersion,
                    }: ModuleRouteContext) =>
                      createCrumb(
                        `/module/${namespace}/${name}/${target}/${rawVersion}`,
                        `${namespace}/${name}`,
                      ),
                  },
                  children: [
                    {
                      element: <Module />,
                      handle: {
                        middleware: moduleMetadataMiddleware,
                      },
                      children: [
                        {
                          index: true,
                          element: <ModuleReadme />,
                          loader: moduleReadmeLoader,
                        },
                        {
                          path: "inputs",
                          element: <ModuleInputs />,
                        },
                        {
                          path: "outputs",
                          element: <ModuleOutputs />,
                        },
                        {
                          path: "dependencies",
                          element: <ModuleDependencies />,
                        },
                        {
                          path: "resources",
                          element: <ModuleResources />,
                        },
                      ],
                    },
                    {
                      path: "example/:example",
                      element: <ModuleExample />,
                      loader: moduleExampleLoader,
                      handle: {
                        middleware: moduleExampleMiddleware,
                        crumb: ({
                          namespace,
                          name,
                          target,
                          example,
                          rawVersion,
                        }: ModuleRouteContext & ModuleExampleRouteContext) =>
                          createCrumb(
                            `/module/${namespace}/${name}/${target}/${rawVersion}/example/${example}`,
                            example,
                          ),
                      },
                      children: [
                        {
                          index: true,
                          element: <ModuleExampleReadme />,
                          loader: moduleExampleReadmeLoader,
                        },
                        {
                          path: "inputs",
                          element: <ModuleExampleInputs />,
                        },
                        {
                          path: "outputs",
                          element: <ModuleExampleOutputs />,
                        },
                      ],
                    },
                    {
                      path: "submodule/:submodule",
                      element: <ModuleSubmodule />,
                      loader: moduleSubmoduleLoader,
                      handle: {
                        middleware: moduleSubmoduleMiddleware,
                        crumb: ({
                          namespace,
                          name,
                          target,
                          submodule,
                          rawVersion,
                        }: ModuleRouteContext & ModuleSubmoduleRouteContext) =>
                          createCrumb(
                            `/module/${namespace}/${name}/${target}/${rawVersion}/submodule/${submodule}`,
                            submodule,
                          ),
                      },
                      children: [
                        {
                          index: true,
                          element: <ModuleSubmoduleReadme />,
                          loader: moduleSubmoduleReadmeLoader,
                        },
                        {
                          path: "inputs",
                          element: <ModuleSubmoduleInputs />,
                        },
                        {
                          path: "outputs",
                          element: <ModuleSubmoduleOutputs />,
                        },
                        {
                          path: "dependencies",
                          element: <ModuleSubmoduleDependencies />,
                        },
                        {
                          path: "resources",
                          element: <ModuleSubmoduleResources />,
                        },
                      ],
                    },
                  ],
                },
              ],
            },
          ],
        },
        {
          id: "provider",
          path: "/provider",
          handle: {
            crumb: () => createCrumb("/providers", "Providers"),
          },
          children: [
            {
              path: ":namespace",
              children: [
                {
                  index: true,
                  element: <Navigate to="/providers" />,
                },
                {
                  path: ":provider/:version?",
                  element: <Provider />,
                  loader: providerLoader,
                  errorElement: <ProviderError />,
                  handle: {
                    middleware: providerMiddleware,
                    crumb: ({
                      namespace,
                      provider,
                      version,
                    }: ProviderRouteContext) =>
                      createCrumb(
                        `/provider/${namespace}/${provider}/${version}`,
                        `${namespace}/${provider}`,
                      ),
                  },
                  children: [
                    {
                      index: true,
                      element: <ProviderOverview />,
                      loader: providerOverviewLoader,
                    },
                    {
                      path: "docs/:type/:doc",
                      element: <ProviderDocs />,
                      loader: providerDocsLoader,
                    },
                  ],
                },
              ],
            },
          ],
        },
      ],
    },
  ],
  {
    async unstable_dataStrategy({ request, params, matches }) {
      const context = {};

      let response: Response | undefined;

      for (const match of matches) {
        if (match.route.handle?.middleware) {
          const result = await match.route.handle.middleware(
            { request, params },
            context,
          );

          if (result) {
            response = result;
          }
        }
      }

      const matchesToLoad = matches.filter((m) => m.shouldLoad);

      const results = await Promise.all(
        matchesToLoad.map((match) =>
          match.resolve(async (handler) => {
            if (response) {
              return response;
            }

            let result = await handler(context);

            if (result && "data" in result) {
              result.data = result.data
                ? {
                    ...result.data,
                    ...context,
                  }
                : context;
            } else if (result) {
              result = {
                ...result,
                ...context,
              };
            }

            return result;
          }),
        ),
      );

      return results.reduce(
        (acc, result, i) =>
          Object.assign(acc, {
            [matchesToLoad[i].route.id]: result,
          }),
        {},
      );
    },
  },
);

if (import.meta.hot) {
  import.meta.hot.dispose(() => router.dispose());
}
