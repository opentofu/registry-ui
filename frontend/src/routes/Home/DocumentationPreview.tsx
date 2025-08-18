import { useQuery } from "@tanstack/react-query";
import { definitions } from "@/api";
import { Markdown } from "@/components/Markdown";
import { api } from "@/query";
import { queryOptions } from "@tanstack/react-query";
import { Link } from "react-router";
import { Button } from "@/components/Button";
import { getDocumentationUrl } from "./utils/getDocumentationUrl";

interface DocumentationPreviewProps {
  result: definitions["SearchResultItem"];
}

const getProviderDocQuery = (
  namespace: string,
  provider: string,
  version: string,
  type?: string,
  name?: string,
) => {
  return queryOptions({
    queryKey: [
      "preview-provider-doc",
      namespace,
      provider,
      type,
      name,
      version,
    ],
    queryFn: async () => {
      const urlBase = `providers/${namespace}/${provider}/${version}/`;
      const path = type && name ? `${type}/${name}.md` : `index.md`;

      try {
        const response = await api(urlBase + path);
        const text = await response.text();
        return text;
      } catch (error) {
        if (error instanceof Error && error.message.includes("404")) {
          return null;
        }
        throw error;
      }
    },
  });
};

const getModuleDocQuery = (
  namespace: string,
  name: string,
  targetSystem: string,
  version: string,
) => {
  return queryOptions({
    queryKey: ["preview-module-doc", namespace, name, targetSystem, version],
    queryFn: async () => {
      const url = `modules/${namespace}/${name}/${targetSystem}/${version}/README.md`;

      try {
        const response = await api(url);
        const text = await response.text();
        return text;
      } catch (error) {
        if (error instanceof Error && error.message.includes("404")) {
          return null;
        }
        throw error;
      }
    },
  });
};

export function DocumentationPreview({ result }: DocumentationPreviewProps) {
  // Determine which query to use based on result type
  const docQuery = (() => {
    const vars = result.link_variables;

    if (result.type === "module") {
      return getModuleDocQuery(
        vars.namespace,
        vars.name,
        vars.target_system,
        vars.version || "latest",
      );
    } else if (result.type === "provider") {
      return getProviderDocQuery(
        vars.namespace,
        vars.name,
        vars.version || "latest",
      );
    } else if (result.type.startsWith("provider/")) {
      const docType = result.type.split("/")[1] + "s"; // resource -> resources
      return getProviderDocQuery(
        vars.namespace,
        vars.name,
        vars.version || "latest",
        docType,
        vars.id,
      );
    }

    return null;
  })();

  const {
    data: content,
    isLoading,
    error,
  } = useQuery({
    ...docQuery!,
    enabled: !!docQuery,
  });

  const fullDocLink = getDocumentationUrl(result);

  if (isLoading) {
    return <DocumentationSkeleton />;
  }

  if (error || !content) {
    return (
      <div className="flex flex-col gap-5 px-5">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {result.link_variables.namespace}/{result.link_variables.name}
            {result.type !== "provider" && result.type !== "module" && (
              <span className="ml-2 text-gray-500 dark:text-gray-400">
                → {result.link_variables.id}
              </span>
            )}
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            {result.type === "provider"
              ? "Provider"
              : result.type === "provider/resource"
                ? "Resource"
                : result.type === "provider/datasource"
                  ? "Data Source"
                  : result.type === "provider/function"
                    ? "Function"
                    : result.type === "module"
                      ? "Module"
                      : result.type}
            {result.link_variables.target_system && (
              <span> • {result.link_variables.target_system}</span>
            )}
          </p>
        </div>

        <div className="p-5">
          <div className="rounded-lg border border-gray-200 bg-white p-8 text-center dark:border-gray-700 dark:bg-gray-800">
            <svg
              className="mx-auto mb-4 h-12 w-12 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p className="mb-4 text-gray-600 dark:text-gray-400">
              Unable to load documentation preview
            </p>
            <Link to={fullDocLink}>
              <Button variant="primary">View Full Documentation →</Button>
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col gap-5 px-5">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {result.link_variables.namespace}/{result.link_variables.name}
            {result.type !== "provider" && result.type !== "module" && (
              <span className="ml-2 text-gray-500 dark:text-gray-400">
                → {result.link_variables.id}
              </span>
            )}
          </h1>
          <div className="flex items-center gap-4">
            <p className="text-gray-600 dark:text-gray-400">
              {result.type === "provider"
                ? "Provider"
                : result.type === "provider/resource"
                  ? "Resource"
                  : result.type === "provider/datasource"
                    ? "Data Source"
                    : result.type === "provider/function"
                      ? "Function"
                      : result.type === "module"
                        ? "Module"
                        : result.type}
              {result.link_variables.target_system && (
                <span> • {result.link_variables.target_system}</span>
              )}
            </p>
            <Link to={fullDocLink}>
              <Button variant="secondary">View Full Docs →</Button>
            </Link>
          </div>
        </div>
      </div>

      <div className="p-5">
        <div className="prose prose-sm dark:prose-invert max-w-none">
          <Markdown text={content} />
        </div>
      </div>
    </>
  );
}

function DocumentationSkeleton() {
  return (
    <>
      <div className="flex flex-col gap-5 px-5">
        <div className="space-y-2">
          <div className="h-9 w-96 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
          <div className="h-5 w-48 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
        </div>
      </div>

      <div className="p-5">
        <div className="space-y-4">
          <div className="h-8 w-64 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
          <div className="space-y-2">
            <div className="h-4 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
            <div className="h-4 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
            <div className="h-4 w-3/4 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
          </div>
          <div className="mt-4 h-32 animate-pulse rounded bg-gray-100 dark:bg-gray-800"></div>
          <div className="space-y-2">
            <div className="h-4 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
            <div className="h-4 w-5/6 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
          </div>
        </div>
      </div>
    </>
  );
}
