import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { useEffect } from "react";
import { getProviderDoc } from "../utils/getProviderDoc";

export function ProviderPageTitle() {
  const { namespace, provider, version, doc, type } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  useEffect(() => {
    const providerDoc = getProviderDoc(data.docs, type, doc);

    if (providerDoc) {
      document.title = `${providerDoc.title} - OpenTofu Registry`;
    }

    return () => {
      document.title = "OpenTofu Registry";
    };
  }, [data, doc, type]);

  return null;
}
