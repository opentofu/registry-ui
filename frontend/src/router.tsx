import { createBrowserRouter, defer, redirect } from "react-router-dom";
import { Home } from "./routes/Home";
import { Modules } from "./routes/Modules";
import { Providers } from "./routes/Providers";

import { createCrumb } from "./crumbs";
import { queryClient } from "./query";
import { Error } from "./routes/Error";
import { Module } from "./routes/Module";
import { ModuleDependencies } from "./routes/Module/Dependencies";
import { ModuleInputs } from "./routes/Module/Inputs";
import { ModuleOutputs } from "./routes/Module/Outputs";
import {
  getModuleDataQuery,
  getModuleReadmeQuery,
  getModuleVersionDataQuery,
} from "./routes/Module/query";
import { ModuleReadme } from "./routes/Module/Readme";
import { ModuleResources } from "./routes/Module/Resources";
import { ModuleExample } from "./routes/ModuleExample";
import { ModuleExampleInputs } from "./routes/ModuleExample/Inputs";
import { ModuleExampleOutputs } from "./routes/ModuleExample/Outputs";
import { ModuleExampleReadme } from "./routes/ModuleExample/Readme";
import { ModuleSubmodule } from "./routes/ModuleSubmodule";
import { ModuleSubmoduleDependencies } from "./routes/ModuleSubmodule/dependencies";
import { ModuleSubmoduleInputs } from "./routes/ModuleSubmodule/inputs";
import { ModuleSubmoduleOutputs } from "./routes/ModuleSubmodule/outputs";
import { ModuleSubmoduleReadme } from "./routes/ModuleSubmodule/readme";
import { ModuleSubmoduleResources } from "./routes/ModuleSubmodule/resources";
import { Provider } from "./routes/Provider";
import { ProviderDocs } from "./routes/Provider/Docs";
import { ProviderOverview } from "./routes/Provider/Overview";
import {
  namespaceProvidersLoader,
  providersLoader,
} from "./routes/Providers/loader";
import { providerOverviewLoader } from "./routes/Provider/Overview/loader";
import { providerDocsLoader } from "./routes/Provider/Docs/loader";
import { providerLoader } from "./routes/Provider/loader";
import { modulesLoader } from "./routes/Modules/loader";
import { getModuleExampleReadmeQuery } from "./routes/ModuleExample/query";
import { providerMiddleware } from "./routes/Provider/middleware";

export const router = createBrowserRouter(
  [
    {
      id: "home",
      index: true,
      element: <Home />,
      errorElement: <Error />,
    },
    {
      id: "providers",
      path: "/providers",
      element: <Providers />,
      errorElement: <Error />,
      loader: providersLoader,
      handle: {
        crumb: () => createCrumb("/providers", "Providers"),
      },
    },
    {
      id: "modules",
      path: "/modules",
      element: <Modules />,
      errorElement: <Error />,
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
      errorElement: <Error />,
      children: [
        {
          path: ":namespace",
          children: [
            {
              index: true,
              element: <Modules />,
              loader: async ({ params }) => {
                return {
                  namespace: params.namespace,
                };
              },
              handle: {
                crumb: ({ namespace }) =>
                  createCrumb(`/module/${namespace}`, namespace),
              },
            },
            {
              path: ":name/:target/:version?",
              element: <Module />,
              loader: async ({ params }) => {
                const data = queryClient.ensureQueryData(
                  getModuleDataQuery(
                    params.namespace,
                    params.name,
                    params.target,
                  ),
                );

                const versionData = queryClient.ensureQueryData(
                  getModuleVersionDataQuery(
                    params.namespace,
                    params.name,
                    params.target,
                    params.version,
                  ),
                );

                return defer({
                  data,
                  versionData,
                  namespace: params.namespace,
                  name: params.name,
                  target: params.target,
                });
              },
              handle: {
                crumb: ({ namespace, name, target }) =>
                  createCrumb(
                    `/module/${namespace}/${name}/${target}`,
                    `${namespace}/${target}`,
                  ),
              },
              children: [
                {
                  index: true,
                  element: <ModuleReadme />,
                  loader: async ({ params }) => {
                    const readme = queryClient.ensureQueryData(
                      getModuleReadmeQuery(
                        params.namespace,
                        params.module,
                        params.provider,
                        params.version,
                      ),
                    );

                    return defer({
                      readme,
                    });
                  },
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
              path: ":name/:target/:version?/example/:example",
              element: <ModuleExample />,
              loader: async ({ params }) => {
                const data = queryClient.ensureQueryData(
                  getModuleDataQuery(
                    params.namespace,
                    params.name,
                    params.target,
                  ),
                );

                return defer({
                  data,

                  namespace: params.namespace,
                  name: params.name,
                  target: params.target,
                  example: params.example,
                });
              },
              handle: {
                crumb: ({ namespace, name, target, example }) => [
                  createCrumb(
                    `/module/${namespace}/${name}/${target}`,
                    `${namespace}/${name}`,
                  ),
                  createCrumb(
                    `/module/${namespace}/${name}/${target}/example/${example}`,
                    example,
                  ),
                ],
              },
              children: [
                {
                  index: true,
                  element: <ModuleExampleReadme />,
                  loader: async ({ params }) => {
                    const readme = queryClient.ensureQueryData(
                      getModuleExampleReadmeQuery(
                        params.namespace,
                        params.module,
                        params.provider,
                        params.version,
                        params.example,
                      ),
                    );

                    return defer({
                      readme,
                    });
                  },
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
              path: ":module/:provider/:version?/submodule/:submodule",
              element: <ModuleSubmodule />,
              loader: async ({ params }) => {
                const data = queryClient.ensureQueryData(
                  getModuleDataQuery(
                    params.namespace,
                    params.module,
                    params.provider,
                    "5.51.1",
                  ),
                );

                return defer({
                  data,

                  namespace: params.namespace,
                  module: params.module,
                  provider: params.provider,
                  submodule: params.submodule,
                });
              },
              handle: {
                crumb: ({ namespace, module, provider, submodule }) => [
                  createCrumb(
                    `/module/${namespace}/${module}/${provider}`,
                    `${namespace}/${module}`,
                  ),
                  createCrumb(
                    `/module/${namespace}/${module}/${provider}/submodule/${submodule}`,
                    submodule,
                  ),
                ],
              },
              children: [
                {
                  index: true,
                  element: <ModuleSubmoduleReadme />,
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
    {
      id: "provider",
      path: "/provider",
      handle: {
        crumb: () => createCrumb("/providers", "Providers"),
      },
      errorElement: <Error />,
      children: [
        {
          path: ":namespace",
          children: [
            {
              index: true,
              element: <Providers />,
              loader: namespaceProvidersLoader,
              handle: {
                crumb: ({ namespace }) =>
                  createCrumb(`/provider/${namespace}`, namespace),
              },
            },
            {
              path: ":provider/:version?",
              element: <Provider />,
              loader: providerLoader,
              handle: {
                middleware: providerMiddleware,
                crumb: ({ namespace, provider, version }) => [
                  createCrumb(
                    `/provider/${namespace}/${provider}/${version}`,
                    `${namespace}/${provider}`,
                  ),
                ],
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

      return Promise.all(
        matches.map((match) => {
          return match.resolve(async (handler) => {
            if (response) {
              return { type: "data", result: response };
            }

            const result = await handler(context);

            if (result && "data" in result) {
              result.data = result.data
                ? {
                    ...result.data,
                    ...context,
                  }
                : context;
            }

            return { type: "data", result };
          });
        }),
      );
    },
  },
);

if (import.meta.hot) {
  import.meta.hot.dispose(() => router.dispose());
}
