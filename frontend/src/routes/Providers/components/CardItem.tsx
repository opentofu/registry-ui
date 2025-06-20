import {
  CardItemFooter,
  CardItemFooterDetail,
  CardItemFooterDetailSkeleton,
} from "@/components/CardItem/Footer";

import { CardItem } from "@/components/CardItem";
import { CardItemTitle } from "@/components/CardItem/Title";
import { DateTime } from "@/components/DateTime";
import { Paragraph } from "@/components/Paragraph";
import { definitions } from "@/api";
import { Link } from "react-router";

interface ProviderCardItemProps {
  provider: definitions["Provider"];
}

export function ProvidersCardItem({ provider }: ProviderCardItemProps) {
  const latestVersion = provider.versions?.[0];

  return (
    <CardItem>
      <div className="flex items-start gap-4 h-full">
        <Link 
          to={`/provider/${provider.addr.namespace}/${provider.addr.name}/${latestVersion?.id || 'latest'}`}
          className="flex-shrink-0"
        >
          <img 
            src={`https://avatars.githubusercontent.com/${provider.addr.namespace}`} 
            alt={`${provider.addr.namespace} avatar`}
            className="w-10 h-10 rounded-lg ring-1 ring-gray-200 dark:ring-gray-700"
            loading="lazy"
            onError={(e) => {
              e.currentTarget.src = '/favicon.ico'; // Fallback to OpenTofu icon
            }}
          />
        </Link>
        
        <div className="flex-grow min-w-0 flex flex-col">
          <CardItemTitle
            linkProps={{
              to: `/provider/${provider.addr.namespace}/${provider.addr.name}/${latestVersion?.id || 'latest'}`,
            }}
            className="line-clamp-1"
          >
            {provider.addr.namespace}/{provider.addr.name}
          </CardItemTitle>

          <Paragraph className="mb-2 mt-1 line-clamp-1 text-sm flex-grow min-h-[1.25rem]">
            {provider.description || '\u00A0'}
          </Paragraph>

          <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-gray-600 dark:text-gray-300 mt-auto">
            {provider.popularity !== undefined && provider.popularity > 0 && (
              <span className="flex items-center gap-1">
                <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.751.751 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25Z"/>
                </svg>
                {provider.popularity.toLocaleString()} stars
              </span>
            )}
            
            {provider.fork_count !== undefined && provider.fork_count > 0 && (
              <span className="flex items-center gap-1">
                <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75v-.878a2.25 2.25 0 1 1 1.5 0v.878a2.25 2.25 0 0 1-2.25 2.25h-1.5v2.128a2.251 2.251 0 1 1-1.5 0V8.5h-1.5A2.25 2.25 0 0 1 3.5 6.25v-.878a2.25 2.25 0 1 1 1.5 0ZM5 3.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Zm6.75.75a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm-3 8.75a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Z"/>
                </svg>
                {provider.fork_count.toLocaleString()} forks
              </span>
            )}

            {provider.fork_of && (
              <span className="flex items-center gap-1">
                <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75v-.878a2.25 2.25 0 1 1 1.5 0v.878a2.25 2.25 0 0 1-2.25 2.25h-1.5v2.128a2.251 2.251 0 1 1-1.5 0V8.5h-1.5A2.25 2.25 0 0 1 3.5 6.25v-.878a2.25 2.25 0 1 1 1.5 0ZM5 3.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Zm6.75.75a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm-3 8.75a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Z"/>
                </svg>
                Forked
              </span>
            )}

            {latestVersion?.id && (
              <span className="flex items-center gap-1">
                {latestVersion.id}
              </span>
            )}

            {latestVersion?.published && (
              <span className="flex items-center gap-1">
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <DateTime value={latestVersion?.published} />
              </span>
            )}
          </div>
        </div>
      </div>
    </CardItem>
  );
}

export function ProvidersCardItemSkeleton() {
  return (
    <CardItem>
      <div className="flex items-start gap-4">
        <div className="w-10 h-10 rounded-lg animate-pulse bg-gray-500/25" />
        
        <div className="flex-grow">
          <span className="flex h-em w-48 animate-pulse bg-gray-500/25 text-xl" />

          <span className="mb-7 mt-5 flex h-em w-96 animate-pulse bg-gray-500/25" />

          <CardItemFooter>
            <CardItemFooterDetailSkeleton />
            <CardItemFooterDetailSkeleton />
          </CardItemFooter>
        </div>
      </div>
    </CardItem>
  );
}
