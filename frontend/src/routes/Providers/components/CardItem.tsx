import { CardItem } from "@/components/CardItem";
import { Paragraph } from "@/components/Paragraph";
import {
  CardItemFooter,
  CardItemFooterDetail,
} from "@/components/CardItem/Footer";
import { CardItemTitle } from "@/components/CardItem/Title";
import { formatDate } from "@/utils/formatDate";
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
          {formatDate(latestVersion.published)}
        </CardItemFooterDetail>
      </CardItemFooter>
    </CardItem>
  );
}

export function ProvidersCardItemSkeleton() {
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
