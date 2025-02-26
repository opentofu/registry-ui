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

interface ProviderCardItemProps {
  addr: definitions["ProviderAddr"];
  description: string;
  latestVersion: definitions["ProviderVersionDescriptor"];
}

export function ProvidersCardItem({
  addr,
  description,
  latestVersion,
}: ProviderCardItemProps) {
  return (
    <CardItem>
      <CardItemTitle
        linkProps={{
          to: `/provider/${addr.namespace}/${addr.name}/${latestVersion.id}`,
        }}
      >
        {addr.namespace}/{addr.name}
      </CardItemTitle>

      <Paragraph className="mb-4 mt-2">{description}</Paragraph>

      <CardItemFooter>
        <CardItemFooterDetail label="Latest version">
          {latestVersion.id}
        </CardItemFooterDetail>
        <CardItemFooterDetail label="Published">
          <DateTime value={latestVersion.published} />
        </CardItemFooterDetail>
      </CardItemFooter>
    </CardItem>
  );
}

export function ProvidersCardItemSkeleton() {
  return (
    <CardItem>
      <span className="flex h-em w-48 animate-pulse bg-gray-500/25 text-xl" />

      <span className="mb-7 mt-5 flex h-em w-96 animate-pulse bg-gray-500/25" />

      <CardItemFooter>
        <CardItemFooterDetailSkeleton />
        <CardItemFooterDetailSkeleton />
      </CardItemFooter>
    </CardItem>
  );
}
