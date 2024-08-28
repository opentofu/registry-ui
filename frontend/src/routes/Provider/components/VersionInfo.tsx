import { useSearchParams } from "react-router-dom";
import { OldVersionBanner } from "@/components/OldVersionBanner";
import {
  LanguagePicker,
  LanguagePickerSkeleton,
} from "@/components/LanguagePicker";
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

  const [, setSearchParams] = useSearchParams();

  const langs = Object.keys(versionData.cdktf_docs);
  const language = langs.includes(lang) ? lang : null;

  const languages = langs.map((language) => ({
    code: language,
    name: languageLabels[language],
  }));

  const handleLanguageChange = (language: string | null) => {
    setSearchParams((params) => {
      const next = new URLSearchParams(params);
      if (language) {
        next.set("lang", language);
      } else {
        next.delete("lang");
      }
      return next;
    });
  };

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
        <LanguagePicker
          languages={languages}
          selected={language}
          onChange={handleLanguageChange}
        />
      </div>
      {version !== latestVersion && (
        <OldVersionBanner latestVersionLink={latestVersionLink} />
      )}
    </div>
  );
}

export function ProviderVersionInfoSkeleton() {
  return (
    <div className="flex items-center justify-between">
      <VersionInfoSkeleton />
      <LanguagePickerSkeleton />
    </div>
  );
}
