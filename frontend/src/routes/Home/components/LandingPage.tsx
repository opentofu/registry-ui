import { Paragraph } from "@/components/Paragraph";

import Logo from "@/logo.svg?react";

interface LandingPageProps {
  query: string;
  onQueryChange: (value: string) => void;
  onSearchFocus: () => void;
  inputRef: React.RefObject<HTMLInputElement | null>;
}

export function LandingPage({
  query,
  onQueryChange,
  onSearchFocus,
  inputRef,
}: LandingPageProps) {
  return (
    <main className="min-h-screen">
      <div className="container m-auto flex flex-col items-center pt-24 text-center">
        <Logo />
        <h2 className="mt-5 max-w-4xl text-5xl leading-tight font-bold text-balance lg:text-6xl">
          Documentation for OpenTofu Providers and Modules
        </h2>
        <Paragraph className="mt-5 mb-7 text-balance">
          This technology preview contains documentation for a select few
          providers, namespaces, and modules in the OpenTofu registry.
          <br />
          <strong>Note:</strong> the data in this interface may not be up to
          date during the beta phase.
        </Paragraph>

        <div className="w-full max-w-xl">
          <div className="relative">
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => onQueryChange(e.target.value)}
              onFocus={() => {
                if (query.length > 0) {
                  onSearchFocus();
                }
              }}
              placeholder="Search providers, resources, or modules (Press / to focus)"
              className="focus:ring-brand-500 h-14 w-full rounded-xl border border-gray-200 bg-white pr-4 pl-12 text-base shadow-sm placeholder:text-gray-400 focus:border-transparent focus:ring-2 focus:outline-none dark:border-gray-700 dark:bg-blue-900 dark:text-gray-200 dark:placeholder-gray-400"
            />
            <div className="absolute top-1/2 left-4 -translate-y-1/2 text-gray-400">
              <svg
                className="size-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="m21 21-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
