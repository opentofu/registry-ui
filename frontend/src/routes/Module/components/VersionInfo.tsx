import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";

export function ModuleVersionInfo() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  const latestVersion = data.versions[0].id;

  const latestVersionLink = `/module/${namespace}/${name}/${target}/${latestVersion}`;

  return (
    <div className="flex items-center justify-between">
      {version !== latestVersion ? (
        <div className="flex items-center gap-3 px-4 py-2 bg-blue-50 dark:bg-blue-950/50 rounded-lg border border-blue-200 dark:border-blue-800">
          <div className="flex items-center gap-2">
            <span className="text-sm text-blue-700 dark:text-blue-300">
              Viewing version {version}
            </span>
            <span className="text-blue-400 dark:text-blue-600">•</span>
            <a 
              href={latestVersionLink}
              className="text-sm font-medium text-blue-700 dark:text-blue-300 hover:text-blue-900 dark:hover:text-blue-100 transition-colors"
            >
              Switch to latest ({latestVersion})
            </a>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2 px-4 py-2 bg-green-50 dark:bg-green-950/50 rounded-lg border border-green-200 dark:border-green-800">
          <span className="text-sm text-green-700 dark:text-green-300">
            ✓ Latest version
          </span>
        </div>
      )}
    </div>
  );
}

export function ModuleVersionInfoSkeleton() {
  return (
    <div className="flex items-center justify-between">
      <span className="h-10 w-64 animate-pulse bg-gray-500/25 rounded-lg" />
    </div>
  );
}
