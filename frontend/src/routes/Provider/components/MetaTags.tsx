import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderDoc } from "../utils/getProviderDoc";
import { MetaTags } from "@/components/MetaTags";
import { Suspense } from "react";

export function ProviderMetaTagsContent() {
  const { namespace, provider, version, doc, type, isLatest, lang } =
    useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  const providerDoc = getProviderDoc(data, type, doc, lang);

  if (!providerDoc) {
    return null;
  }

  let title = `${namespace}/${provider}`;

  if (!isLatest) {
    title = `${version} - ${title}`;
  }

  return (
    <MetaTags
      title={`${providerDoc.title} - ${title}`}
      description={providerDoc.description}
    />
  );
}

export function ProviderMetaTags() {
  return (
    <Suspense fallback={null}>
      <ProviderMetaTagsContent />
    </Suspense>
  );
}
