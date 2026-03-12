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

interface ModulesCardItemProps {
  addr: definitions["ModuleAddr"];
  description: string;
  latestVersion: definitions["ModuleVersionDescriptor"];
}

export function ModulesCardItem({
  addr,
  description,
  latestVersion,
}: ModulesCardItemProps) {
  return (
    <CardItem>
      <CardItemTitle
        linkProps={{
          to: `/module/${addr.namespace}/${addr.name}/${addr.target}/latest`,
        }}
      >
        {addr.namespace}/{addr.name}
      </CardItemTitle>

      <Paragraph className="mt-2 mb-4">{description}</Paragraph>

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

export function ModulesCardItemSkeleton() {
  return (
    <CardItem>
      <span className="h-em flex w-48 animate-pulse bg-gray-500/25 text-xl" />

      <span className="h-em mt-5 mb-7 flex w-96 animate-pulse bg-gray-500/25" />

      <CardItemFooter>
        <CardItemFooterDetailSkeleton />
        <CardItemFooterDetailSkeleton />
      </CardItemFooter>
    </CardItem>
  );
}
