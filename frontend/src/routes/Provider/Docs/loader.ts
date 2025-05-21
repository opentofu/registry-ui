import { queryClient } from "@/query";
import { isValidDocsType } from "../utils/isValidDocsType";
import { getProviderDocsQuery } from "../query";
import { LoaderFunction } from "react-router";
import { ProviderRouteContext } from "../types";

export const providerDocsLoader: LoaderFunction = (
  { params, request },
  context,
) => {
  if (!isValidDocsType(params.type)) {
    throw new Error("Invalid doc type");
  }

  const url = new URL(request.url);
  const lang = url.searchParams.get("lang");

  const docs = queryClient.ensureQueryData(
    getProviderDocsQuery(
      params.namespace,
      params.provider,
      (context as ProviderRouteContext).version,
      params.type,
      params.doc,
      lang,
    ),
  );

  return {
    docs,
  };
};
