import { LanguagePicker } from "@/components/LanguagePicker";
import { useSuspenseQueries } from "@tanstack/react-query";
import { getProviderDataQuery, getProviderVersionDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";
import { isValidDocsType } from "../utils/isValidDocsType";

const languageLabels: { [key: string]: string } = {
  typescript: "TypeScript",
  python: "Python",
  go: "Go",
  java: "Java",
  csharp: "C#",
};

export function ProviderVersionInfo() {
  const { namespace, provider, version, lang, type, doc } = useProviderParams();

  const [{ data: versionData }, { data }] = useSuspenseQueries({
    queries: [
      getProviderVersionDataQuery(namespace, provider, version),
      getProviderDataQuery(namespace, provider),
    ],
  });

  const langs = Object.keys(versionData.cdktf_docs);

  const languages = langs.map((language) => ({
    code: language,
    name: languageLabels[language],
  }));

  const latestVersion = data.versions[0].id;

  let latestVersionLink = `/provider/${namespace}/${provider}/${latestVersion}`;

  if (isValidDocsType(type) && doc) {
    latestVersionLink += `/docs/${type}/${doc}`;
  }

  if (lang) {
    latestVersionLink += `?lang=${lang}`;
  }

  return (
    <div className="flex items-center justify-between">
      {version !== latestVersion ? (
        <div className="flex items-center gap-3 px-4 py-2 bg-blue-50 dark:bg-blue-950/50 rounded-lg border border-blue-200 dark:border-blue-800">
          <div className="flex items-center gap-2">
            <span className="text-sm text-blue-700 dark:text-blue-300">
              Viewing version {version}
            </span>
            <span className="text-blue-400 dark:text-blue-600">•</span>
            <a 
              href={latestVersionLink}
              className="text-sm font-medium text-blue-700 dark:text-blue-300 hover:text-blue-900 dark:hover:text-blue-100 transition-colors"
            >
              Switch to latest ({latestVersion})
            </a>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2 px-4 py-2 bg-green-50 dark:bg-green-950/50 rounded-lg border border-green-200 dark:border-green-800">
          <span className="text-sm text-green-700 dark:text-green-300">
            ✓ Latest version
          </span>
        </div>
      )}
      {languages.length > 0 && <LanguagePicker languages={languages} />}
    </div>
  );
}

export function ProviderVersionInfoSkeleton() {
  return (
    <div className="flex items-center justify-between">
      <span className="h-10 w-64 animate-pulse bg-gray-500/25 rounded-lg" />
      <span className="h-10 w-48 animate-pulse bg-gray-500/25 rounded" />
    </div>
  );
}
