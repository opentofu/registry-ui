import { Suspense } from "react";
import {
  ProviderDocsContent,
  ProviderDocsContentSkeleton,
} from "../components/DocsContent";

export function ProviderOverview() {
  return (
    <Suspense fallback={<ProviderDocsContentSkeleton />}>
      <ProviderDocsContent />
    </Suspense>
  );
}
