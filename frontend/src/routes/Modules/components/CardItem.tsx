import {
  CardItemFooter,
  CardItemFooterDetail,
  CardItemFooterDetailSkeleton,
} from "@/components/CardItem/Footer";

import { CardItem } from "@/components/CardItem";
import { CardItemTitle } from "@/components/CardItem/Title";
import { Paragraph } from "@/components/Paragraph";
import { definitions } from "@/api";
import { formatDateTag } from "@/utils/formatDate";

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

      <Paragraph className="mb-4 mt-2">{description}</Paragraph>

      <CardItemFooter>
        <CardItemFooterDetail label="Latest version">
          {latestVersion.id}
        </CardItemFooterDetail>
        <CardItemFooterDetail label="Published">
          {formatDateTag(latestVersion.published)}
        </CardItemFooterDetail>
      </CardItemFooter>
    </CardItem>
  );
}

export function ModulesCardItemSkeleton() {
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
