import { PageTitle } from "@/components/PageTitle";
import { Paragraph } from "@/components/Paragraph";
import { InfoSection, InfoSectionItem } from "@/components/InfoSection";
import { Breadcrumbs, BreadcrumbsSkeleton } from "@/components/Breadcrumbs";
import { ReactNode } from "react";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderDataQuery } from "../query";
import { formatDate } from "@/utils/formatDate";
import { useProviderParams } from "../hooks/useProviderParams";

interface WrapperProps {
  children: ReactNode;
}

function Wrapper({ children }: WrapperProps) {
  return (
    <div className="-mx-5 border-b border-gray-200 px-5 pb-5 dark:border-gray-800">
      {children}
    </div>
  );
}

export function ProviderHeader() {
  const { namespace, provider } = useProviderParams();

  const { data } = useSuspenseQuery(getProviderDataQuery(namespace, provider));

  return (
    <Wrapper>
      <Breadcrumbs />
      <div className="mb-5 flex flex-col gap-5">
        <PageTitle>
          {data.addr.namespace}/{data.addr.name}
        </PageTitle>
        {data.description && <Paragraph>{data.description}</Paragraph>}
      </div>
      <InfoSection>
        <InfoSectionItem label="Owner">{data.addr.namespace}</InfoSectionItem>
        <InfoSectionItem label="Latest version">
          {data.versions[0].id}
        </InfoSectionItem>
        <InfoSectionItem label="Published">
          {formatDate(data.versions[0].published)}
        </InfoSectionItem>
      </InfoSection>
    </Wrapper>
  );
}

export function ProviderHeaderSkeleton() {
  return (
    <Wrapper>
      <BreadcrumbsSkeleton />
      <span className="flex h-em w-72 animate-pulse bg-gray-500/25 text-5xl" />
      <span className="mb-6 mt-6 flex h-em w-[450px] animate-pulse bg-gray-500/25" />

      <InfoSection>
        <InfoSectionItem label="Owner">
          <span className="flex h-em w-32 animate-pulse bg-gray-500/25" />
        </InfoSectionItem>
        <InfoSectionItem label="Latest version">
          <span className="flex h-em w-16 animate-pulse bg-gray-500/25" />
        </InfoSectionItem>
        <InfoSectionItem label="Published">
          <span className="flex h-em w-32 animate-pulse bg-gray-500/25" />
        </InfoSectionItem>
      </InfoSection>
    </Wrapper>
  );
}
