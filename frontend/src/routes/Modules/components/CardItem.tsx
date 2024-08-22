import { CardItem } from "@/components/CardItem";
import { Paragraph } from "@/components/Paragraph";
import {
  CardItemFooter,
  CardItemFooterDetail,
} from "@/components/CardItem/Footer";
import { CardItemTitle } from "@/components/CardItem/Title";
import { definitions } from "@/api";
import { formatDate } from "@/utils/formatDate";

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
          {formatDate(latestVersion.published)}
        </CardItemFooterDetail>
      </CardItemFooter>
    </CardItem>
  );
}

export function ModulesCardItemSkeleton() {
  return (
    <CardItem>
      <span className="flex h-em w-48 animate-pulse bg-gray-500/25 text-xl" />

      <span className="mt-5 flex h-em w-96 animate-pulse bg-gray-500/25" />

      <span className="mt-7 flex gap-10">
        <span className="flex h-em w-36 animate-pulse bg-gray-500/25" />
        <span className="flex h-em w-52 animate-pulse bg-gray-500/25" />
      </span>
    </CardItem>
  );
}
