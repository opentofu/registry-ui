import { Markdown } from "@/components/Markdown";
import { useSuspenseQueries } from "@tanstack/react-query";
import { getProviderDocsQuery, getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderDoc } from "../utils/getProviderDoc";
import { EditLink } from "@/components/EditLink";

export function ProviderDocsContent() {
  const { namespace, provider, type, doc, version, lang } = useProviderParams();

  const [{ data: docs }, { data: versionData }] = useSuspenseQueries({
    queries: [
      getProviderDocsQuery(namespace, provider, version, type, doc, lang),
      getProviderVersionDataQuery(namespace, provider, version),
    ],
  });

  const editLink = getProviderDoc(versionData.docs, type, doc)?.edit_link;

  return (
    <>
      <Markdown text={docs} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

export function ProviderDocsContentSkeleton() {
  return (
    <>
      <span className="flex h-em w-52 animate-pulse bg-gray-500/25 text-4xl" />

      <span className="mt-5 flex h-em w-[500px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[400px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[450px] animate-pulse bg-gray-500/25" />

      <span className="mt-8 flex h-em w-[300px] animate-pulse bg-gray-500/25 text-3xl" />

      <span className="mt-5 flex h-em w-[600px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[550px] animate-pulse bg-gray-500/25" />
      <span className="mt-1 flex h-em w-[350px] animate-pulse bg-gray-500/25" />
    </>
  );
}
