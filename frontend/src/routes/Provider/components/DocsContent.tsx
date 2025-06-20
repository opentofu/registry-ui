import { Markdown } from "@/components/Markdown";
import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import {
  DocNotFoundError,
  getProviderDocsQuery,
  getProviderVersionDataQuery,
} from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderDoc } from "../utils/getProviderDoc";
import { EditLink } from "@/components/EditLink";
import { reworkRelativePaths } from "./docsProcessor";
import { useEffect } from "react";
import { useLocation } from "react-router";

export function ProviderDocsContent() {
  const { namespace, provider, type, doc, version, lang } = useProviderParams();
  const location = useLocation();

  const { data: docs, error } = useQuery({
    ...getProviderDocsQuery(namespace, provider, version, type, doc, lang),
    retry: false,
  });

  const { data: versionData } = useSuspenseQuery(
    getProviderVersionDataQuery(namespace, provider, version),
  );

  const editLink = getProviderDoc(versionData, type, doc, lang)?.edit_link;

  // Handle hash scrolling after content loads
  useEffect(() => {
    if (docs && location.hash) {
      // Small delay to ensure DOM is updated
      setTimeout(() => {
        const id = location.hash.slice(1);
        const element = document.getElementById(id);
        if (element) {
          element.scrollIntoView({ behavior: 'smooth', block: 'center' });
          
          // Add flash animation after scroll
          element.style.transition = 'background-color 0.3s ease-in-out';
          element.style.backgroundColor = 'rgb(250 204 21 / 0.3)'; // Yellow with opacity
          
          // Remove the highlight after 2.5 seconds
          setTimeout(() => {
            element.style.backgroundColor = '';
            // Clean up the transition after animation
            setTimeout(() => {
              element.style.transition = '';
            }, 300);
          }, 2500);
        }
      }, 100);
    }
  }, [docs, location.hash]);

  if (error) {
    if (
      error instanceof DocNotFoundError &&
      error.statusCode === 404 &&
      versionData
    ) {
      // handle 404 errors specifically if and only if we have the version data AND we got a 404
      return <ProviderDocsContentNotFound />;
    } else {
      // throw it back up to be handled by the error boundary
      throw error;
    }
  }

  if (!docs) {
    return <ProviderDocsContentSkeleton />;
  }

  let finalDocs = docs;
  if (namespace && provider && version) {
    finalDocs = reworkRelativePaths(docs || "", namespace, provider, version);
  }

  return (
    <>
      <Markdown text={finalDocs} />
      {editLink && <EditLink url={editLink} />}
    </>
  );
}

export function ProviderDocsContentNotFound() {
  return (
    <div className="py-8">
      <h2 className="mb-4 text-2xl font-bold">Documentation not found</h2>
      <p>
        The requested documentation page could not be found, but provider
        information is available.
      </p>
    </div>
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
