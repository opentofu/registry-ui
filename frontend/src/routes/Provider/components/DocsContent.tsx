import { Markdown } from "@/components/Markdown";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderDocsQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";

export function ProviderDocsContent() {
  const { namespace, provider, type, doc, version, lang } = useProviderParams();

  const { data } = useSuspenseQuery(
    getProviderDocsQuery(namespace, provider, version, type, doc, lang),
  );

  return <Markdown text={data} />;
}

export function ProviderDocsContentSkeleton() {
  return (
    <>
      <span className="mt-6 flex h-em w-52 animate-pulse bg-gray-500/25 text-4xl" />

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
