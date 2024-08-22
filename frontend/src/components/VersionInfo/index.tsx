interface VersionInfoProps {
  currentVersion: string;
  latestVersion: string;
}

export function VersionInfo({
  currentVersion,
  latestVersion,
}: VersionInfoProps) {
  return (
    <span className="flex gap-2 text-gray-700 dark:text-gray-300">
      Version displayed
      <strong className="font-normal">
        {currentVersion}
        {currentVersion === latestVersion ? " (latest)" : ""}
      </strong>
    </span>
  );
}

export function VersionInfoSkeleton() {
  return (
    <span className="flex items-center gap-2 text-gray-700 dark:text-gray-300">
      Version displayed
      <span className="flex h-5 w-24 animate-pulse bg-gray-500/25" />
    </span>
  );
}
