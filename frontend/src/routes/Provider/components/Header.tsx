import { DateTime } from "@/components/DateTime";
import { ReactNode } from "react";
import { getProviderDataQuery, getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { useSuspenseQueries } from "@tanstack/react-query";
import { Icon } from "@/components/Icon";
import { github } from "@/icons/github";
import { LicenseInfo } from "@/components/LicenseInfo";

interface WrapperProps {
  children: ReactNode;
}

function Wrapper({ children }: WrapperProps) {
  return (
    <div className="-mx-5 border-b border-gray-200 px-5 pb-5 dark:border-gray-700">
      {children}
    </div>
  );
}

export function ProviderHeader() {
  const { namespace, provider, version } = useProviderParams();

  const [{ data }, { data: versionData }] = useSuspenseQueries({
    queries: [
      getProviderDataQuery(namespace, provider),
      getProviderVersionDataQuery(namespace, provider, version),
    ],
  });

  return (
    <Wrapper>
      <div className="flex flex-col gap-6">
        <div>
          <div className="flex items-center gap-4">
            <img
              src={`https://github.com/${data.addr.namespace}.png`}
              alt={`${data.addr.namespace} avatar`}
              className="h-16 w-16 rounded-lg border border-gray-200 dark:border-gray-700"
            />
            <div>
              <span className="mb-2 text-sm font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400">
                Provider
              </span>
              <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
                {data.addr.namespace}/{data.addr.name}
              </h1>
            </div>
          </div>
          {data.description && (
            <p className="mt-3 text-lg leading-relaxed text-gray-600 dark:text-gray-400">
              {data.description}
            </p>
          )}
        </div>

        <div className="flex gap-8">
          <div>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              Owner
            </span>
            <a
              href={`https://github.com/${data.addr.namespace}`}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-brand-600 dark:hover:text-brand-500 mt-1 block text-base font-medium text-gray-900 transition-colors dark:text-white"
            >
              {data.addr.namespace}
            </a>
          </div>
          <div>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              Published
            </span>
            <p className="mt-1 text-base font-medium text-gray-900 dark:text-white">
              <DateTime value={data.versions[0].published} />
            </p>
          </div>
          {data.popularity > 0 && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                Stars
              </span>
              <p className="mt-1 text-base font-medium text-gray-900 dark:text-white">
                {data.popularity.toLocaleString()}
              </p>
            </div>
          )}
          {data.fork_count > 0 && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                Forks
              </span>
              <p className="mt-1 text-base font-medium text-gray-900 dark:text-white">
                {data.fork_count.toLocaleString()}
              </p>
            </div>
          )}
          {data.fork_of && data.fork_of.namespace && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                Forked from
              </span>
              <a
                href={`/provider/${data.fork_of.namespace}/${data.fork_of.name}/latest`}
                className="hover:text-brand-600 dark:hover:text-brand-500 mt-1 block text-base font-medium text-gray-900 transition-colors dark:text-white"
              >
                {data.fork_of.display ||
                  `${data.fork_of.namespace}/${data.fork_of.name}`}
              </a>
            </div>
          )}
          {versionData.link && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                Repository
              </span>
              <a
                href={versionData.link}
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-brand-600 dark:hover:text-brand-500 mt-1 flex items-center gap-1.5 text-base font-medium text-gray-900 transition-colors dark:text-white"
              >
                <Icon path={github} className="size-4" />
                <span>
                  {
                    versionData.link
                      .replace("https://github.com/", "")
                      .split("/tree/")[0]
                  }
                </span>
              </a>
            </div>
          )}
          <div>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              License
            </span>
            <div className="mt-1">
              <LicenseInfo license={versionData.license} />
            </div>
          </div>
        </div>
      </div>
    </Wrapper>
  );
}

export function ProviderHeaderSkeleton() {
  return (
    <Wrapper>
      <div className="flex flex-col gap-6">
        <div>
          <div className="flex items-center gap-4">
            <span className="h-16 w-16 animate-pulse rounded-lg bg-gray-500/25" />
            <div>
              <span className="mb-2 h-5 w-16 animate-pulse rounded bg-gray-500/25" />
              <span className="flex h-10 w-80 animate-pulse rounded bg-gray-500/25" />
            </div>
          </div>
          <span className="mt-3 flex h-6 w-[450px] animate-pulse rounded bg-gray-500/25" />
        </div>

        <div className="flex gap-8">
          <div>
            <span className="flex h-4 w-12 animate-pulse rounded bg-gray-500/25" />
            <span className="mt-1 flex h-5 w-24 animate-pulse rounded bg-gray-500/25" />
          </div>
          <div>
            <span className="flex h-4 w-16 animate-pulse rounded bg-gray-500/25" />
            <span className="mt-1 flex h-5 w-32 animate-pulse rounded bg-gray-500/25" />
          </div>
          <div>
            <span className="flex h-4 w-12 animate-pulse rounded bg-gray-500/25" />
            <span className="mt-1 flex h-5 w-16 animate-pulse rounded bg-gray-500/25" />
          </div>
          <div>
            <span className="flex h-4 w-12 animate-pulse rounded bg-gray-500/25" />
            <span className="mt-1 flex h-5 w-16 animate-pulse rounded bg-gray-500/25" />
          </div>
          <div>
            <span className="flex h-4 w-12 animate-pulse rounded bg-gray-500/25" />
            <span className="mt-1 flex h-5 w-20 animate-pulse rounded bg-gray-500/25" />
          </div>
        </div>
      </div>
    </Wrapper>
  );
}
