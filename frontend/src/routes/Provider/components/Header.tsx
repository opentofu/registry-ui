import { DateTime } from "@/components/DateTime";
import { ReactNode } from "react";
import { getProviderDataQuery, getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { useSuspenseQueries } from "@tanstack/react-query";
import { Icon } from "@/components/Icon";
import { github } from "@/icons/github";

interface WrapperProps {
  children: ReactNode;
}

function Wrapper({ children }: WrapperProps) {
  return (
    <div className="-mx-5 border-b border-gray-200 dark:border-gray-700 px-5 pb-5">
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
              className="w-16 h-16 rounded-lg border border-gray-200 dark:border-gray-700"
            />
            <div>
              <span className="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                Provider
              </span>
              <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
                {data.addr.namespace}/{data.addr.name}
              </h1>
            </div>
          </div>
          {data.description && (
            <p className="mt-3 text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
              {data.description}
            </p>
          )}
        </div>
        
        <div className="flex gap-8">
          <div>
            <span className="text-sm text-gray-500 dark:text-gray-400">Owner</span>
            <a 
              href={`https://github.com/${data.addr.namespace}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-base font-medium text-gray-900 dark:text-white mt-1 block hover:text-brand-600 dark:hover:text-brand-500 transition-colors"
            >
              {data.addr.namespace}
            </a>
          </div>
          <div>
            <span className="text-sm text-gray-500 dark:text-gray-400">Published</span>
            <p className="text-base font-medium text-gray-900 dark:text-white mt-1">
              <DateTime value={data.versions[0].published} />
            </p>
          </div>
          {data.popularity > 0 && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">Stars</span>
              <p className="text-base font-medium text-gray-900 dark:text-white mt-1">
                {data.popularity.toLocaleString()}
              </p>
            </div>
          )}
          {data.fork_count > 0 && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">Forks</span>
              <p className="text-base font-medium text-gray-900 dark:text-white mt-1">
                {data.fork_count.toLocaleString()}
              </p>
            </div>
          )}
          {data.fork_of && data.fork_of.namespace && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">Forked from</span>
              <a 
                href={`/provider/${data.fork_of.namespace}/${data.fork_of.name}/latest`}
                className="text-base font-medium text-gray-900 dark:text-white mt-1 block hover:text-brand-600 dark:hover:text-brand-500 transition-colors"
              >
                {data.fork_of.display || `${data.fork_of.namespace}/${data.fork_of.name}`}
              </a>
            </div>
          )}
          {versionData.link && (
            <div>
              <span className="text-sm text-gray-500 dark:text-gray-400">Repository</span>
              <a 
                href={versionData.link}
                target="_blank"
                rel="noopener noreferrer"
                className="text-base font-medium text-gray-900 dark:text-white mt-1 flex items-center gap-1.5 hover:text-brand-600 dark:hover:text-brand-500 transition-colors"
              >
                <Icon path={github} className="size-4" />
                <span>{versionData.link.replace('https://github.com/', '').split('/tree/')[0]}</span>
              </a>
            </div>
          )}
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
            <span className="w-16 h-16 animate-pulse bg-gray-500/25 rounded-lg" />
            <div>
              <span className="h-5 w-16 animate-pulse bg-gray-500/25 rounded mb-2" />
              <span className="flex h-10 w-80 animate-pulse bg-gray-500/25 rounded" />
            </div>
          </div>
          <span className="mt-3 flex h-6 w-[450px] animate-pulse bg-gray-500/25 rounded" />
        </div>
        
        <div className="flex gap-8">
          <div>
            <span className="flex h-4 w-12 animate-pulse bg-gray-500/25 rounded" />
            <span className="mt-1 flex h-5 w-24 animate-pulse bg-gray-500/25 rounded" />
          </div>
          <div>
            <span className="flex h-4 w-16 animate-pulse bg-gray-500/25 rounded" />
            <span className="mt-1 flex h-5 w-32 animate-pulse bg-gray-500/25 rounded" />
          </div>
          <div>
            <span className="flex h-4 w-12 animate-pulse bg-gray-500/25 rounded" />
            <span className="mt-1 flex h-5 w-16 animate-pulse bg-gray-500/25 rounded" />
          </div>
          <div>
            <span className="flex h-4 w-12 animate-pulse bg-gray-500/25 rounded" />
            <span className="mt-1 flex h-5 w-16 animate-pulse bg-gray-500/25 rounded" />
          </div>
        </div>
      </div>
    </Wrapper>
  );
}
