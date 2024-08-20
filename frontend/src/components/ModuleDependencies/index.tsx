import { definitions } from "@/api";
import { EmptyState } from "../EmptyState";

interface ModuleDependenciesProps {
  moduleDependencies: Array<definitions["ModuleDependency"]>;
  providerDependencies: Array<definitions["ProviderDependency"]>;
}

export function ModuleDependencies({
  moduleDependencies,
  providerDependencies,
}: ModuleDependenciesProps) {
  return (
    <>
      <div className="border-b border-gray-200 p-5 dark:border-gray-800">
        <h3 className="mb-2 text-3xl font-semibold">Module dependencies</h3>

        {moduleDependencies.length === 0 && (
          <EmptyState
            text="This module has no external module dependencies."
            className="mt-5"
          />
        )}
        {moduleDependencies.length > 0 && (
          <ul className="mt-6 space-y-5">
            {moduleDependencies.map((dependency) => (
              <li key={dependency.name}>
                {dependency.name}{" "}
                {!!dependency.version_constraint && (
                  <span className="text-gray-600 dark:text-gray-500">
                    ({dependency.version_constraint})
                  </span>
                )}
                <br />
                <code className="text-sm text-purple-700 dark:text-purple-300">
                  {dependency.source}
                </code>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="px-5 pt-5">
        <h3 className="mb-2 text-3xl font-semibold">Provider dependencies</h3>

        {providerDependencies.length === 0 && (
          <EmptyState
            text="This module has no provider dependencies."
            className="mt-5"
          />
        )}

        {providerDependencies.length > 0 && (
          <ul className="mt-6 space-y-5">
            {providerDependencies.map((dependency) => (
              <li key={dependency.name}>
                {dependency.name}{" "}
                <span className="text-gray-600 dark:text-gray-500">
                  ({dependency.full_name})
                </span>
                <code className="text-mono text-purple-700 dark:text-purple-300">
                  {dependency.version_constraint}
                </code>
              </li>
            ))}
          </ul>
        )}
      </div>
    </>
  );
}
