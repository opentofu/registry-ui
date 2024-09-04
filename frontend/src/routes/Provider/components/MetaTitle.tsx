import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderDoc } from "../utils/getProviderDoc";
import { MetaTitle } from "@/components/MetaTitle";
import { Suspense } from "react";

export function ProviderMetaTitleContent() {
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

  return <MetaTitle>{`${providerDoc.title} - ${title}`}</MetaTitle>;
}

export function ProviderMetaTitle() {
  return (
    <Suspense fallback={null}>
      <ProviderMetaTitleContent />
    </Suspense>
  );
}
