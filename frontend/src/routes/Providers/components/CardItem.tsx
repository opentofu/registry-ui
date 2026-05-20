import {
  CardItemFooter,
  CardItemFooterDetail,
  CardItemFooterDetailSkeleton,
} from "@/components/CardItem/Footer";

import { CardItem } from "@/components/CardItem";
import { CardItemTitle } from "@/components/CardItem/Title";
import { DateTime } from "@/components/DateTime";
import { Paragraph } from "@/components/Paragraph";
import { Icon } from "@/components/Icon";
import { definitions } from "@/api";
import { Link } from "react-router";
import { star } from "@/icons/star";
import { fork } from "@/icons/fork";
import { clock } from "@/icons/clock";

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
                <Icon path={star} className="w-3 h-3" width={16} height={16} />
                {provider.popularity.toLocaleString()} stars
              </span>
            )}
            
            {provider.fork_count !== undefined && provider.fork_count > 0 && (
              <span className="flex items-center gap-1">
                <Icon path={fork} className="w-3 h-3" width={16} height={16} />
                {provider.fork_count.toLocaleString()} forks
              </span>
            )}

            {provider.fork_of && (
              <span className="flex items-center gap-1">
                <Icon path={fork} className="w-3 h-3" width={16} height={16} />
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
                <Icon path={clock} className="w-3 h-3" width={24} height={24} />
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
