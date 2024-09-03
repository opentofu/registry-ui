import { clsx } from "clsx";
import { Link, useSearchParams } from "react-router-dom";

interface LanguageProps {
  name: string;
  code: string | null;
}

function Language({ name, code }: LanguageProps) {
  const [searchParams] = useSearchParams();
  const isActive = searchParams.get("lang") === code;

  return (
    <Link
      to={{ search: code ? `?lang=${code}` : "" }}
      className={clsx(
        "ml-2 inline-flex h-10 items-center border px-3 font-semibold text-inherit",
        isActive
          ? "border-brand-500 bg-brand-500 dark:border-brand-800 dark:bg-brand-800 dark:text-brand-600"
          : "border-gray-200 dark:border-gray-800",
      )}
    >
      {name}
    </Link>
  );
}

interface LanguagePickerProps {
  languages: Array<{ name: string; code: string }>;
}

export function LanguagePicker({ languages }: LanguagePickerProps) {
  return (
    <nav className="flex items-center">
      <span className="mr-2 text-gray-700 dark:text-gray-300">
        Provider language
      </span>
      <Language name="OpenTofu" code={null} />
      {languages.map(({ name, code }) => (
        <Language key={code} name={name} code={code} />
      ))}
    </nav>
  );
}
