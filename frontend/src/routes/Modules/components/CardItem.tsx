import { CardItem } from "@/components/CardItem";
import { CardItemTitle } from "@/components/CardItem/Title";
import { DateTime } from "@/components/DateTime";
import { Paragraph } from "@/components/Paragraph";
import { Icon } from "@/components/Icon";
import { definitions } from "@/api";
import { Link } from "react-router";
import { star } from "@/icons/star";
import { fork } from "@/icons/fork";
import { target } from "@/icons/target";
import { clock } from "@/icons/clock";

interface ModulesCardItemProps {
  module: definitions["Module"];
}

export function ModulesCardItem({ module }: ModulesCardItemProps) {
  const latestVersion = module.versions?.[0];

  return (
    <CardItem>
      <div className="flex h-full items-start gap-4">
        <Link
          to={`/module/${module.addr.namespace}/${module.addr.name}/${module.addr.target}/latest`}
          className="flex-shrink-0"
        >
          <img
            src={`https://avatars.githubusercontent.com/${module.addr.namespace}`}
            alt={`${module.addr.namespace} avatar`}
            className="h-10 w-10 rounded-lg ring-1 ring-gray-200 dark:ring-gray-700"
            loading="lazy"
            onError={(e) => {
              e.currentTarget.src = "/favicon.ico"; // Fallback to OpenTofu icon
            }}
          />
        </Link>

        <div className="flex min-w-0 flex-grow flex-col">
          <CardItemTitle
            linkProps={{
              to: `/module/${module.addr.namespace}/${module.addr.name}/${module.addr.target}/latest`,
            }}
            className="line-clamp-1"
          >
            {module.addr.display ||
              `${module.addr.namespace}/${module.addr.name}/${module.addr.target}`}
          </CardItemTitle>

          <Paragraph className="mt-1 mb-2 line-clamp-1 min-h-[1.25rem] flex-grow text-sm">
            {module.description || "\u00A0"}
          </Paragraph>

          <div className="mt-auto flex flex-wrap gap-x-3 gap-y-1 text-xs text-gray-600 dark:text-gray-300">
            {module.popularity !== undefined && module.popularity > 0 && (
              <span className="flex items-center gap-1">
                <Icon path={star} className="h-3 w-3" width={16} height={16} />
                {module.popularity.toLocaleString()} stars
              </span>
            )}

            {module.fork_count !== undefined && module.fork_count > 0 && (
              <span className="flex items-center gap-1">
                <Icon path={fork} className="h-3 w-3" width={16} height={16} />
                {module.fork_count.toLocaleString()} forks
              </span>
            )}

            {module.fork_of && (
              <span className="flex items-center gap-1">
                <Icon path={fork} className="h-3 w-3" width={16} height={16} />
                Forked
              </span>
            )}

            <span className="flex items-center gap-1">
              <Icon path={target} className="h-3 w-3" width={24} height={24} />
              {module.addr.target}
            </span>

            {latestVersion?.id && (
              <span className="flex items-center gap-1">
                {latestVersion.id}
              </span>
            )}

            {latestVersion?.published && (
              <span className="flex items-center gap-1">
                <Icon path={clock} className="h-3 w-3" width={24} height={24} />
                <DateTime value={latestVersion?.published} />
              </span>
            )}
          </div>
        </div>
      </div>
    </CardItem>
  );
}

export function ModulesCardItemSkeleton() {
  return (
    <CardItem>
      <div className="flex items-start gap-4">
        <div className="h-10 w-10 animate-pulse rounded-lg bg-gray-500/25" />

        <div className="flex-grow">
          <span className="h-em flex w-48 animate-pulse bg-gray-500/25 text-xl" />

          <span className="h-em mt-5 mb-7 flex w-96 animate-pulse bg-gray-500/25" />

          <div className="flex gap-3">
            <span className="h-em flex w-16 animate-pulse bg-gray-500/25 text-xs" />
            <span className="h-em flex w-16 animate-pulse bg-gray-500/25 text-xs" />
          </div>
        </div>
      </div>
    </CardItem>
  );
}
