import { OldVersionBanner } from "@/components/OldVersionBanner";
import { LanguagePicker } from "@/components/LanguagePicker";
import { VersionInfo, VersionInfoSkeleton } from "@/components/VersionInfo";
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
    <div className="flex flex-col gap-5">
      <div className="flex items-center justify-between">
        <VersionInfo currentVersion={version} latestVersion={latestVersion} />
        {languages.length > 0 && <LanguagePicker languages={languages} />}
      </div>
      {version !== latestVersion && (
        <OldVersionBanner latestVersionLink={latestVersionLink} />
      )}
    </div>
  );
}

export function ProviderVersionInfoSkeleton() {
  return <VersionInfoSkeleton />;
}
