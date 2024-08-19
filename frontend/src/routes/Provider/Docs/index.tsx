import { Suspense } from "react";

import {
  ProviderDocsContent,
  ProviderDocsContentSkeleton,
} from "../components/DocsContent";

export function ProviderDocs() {
  return (
    <Suspense fallback={<ProviderDocsContentSkeleton />}>
      <ProviderDocsContent />
    </Suspense>
  );
}
