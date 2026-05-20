import { clsx } from "clsx";
import { Link, useSearchParams } from "react-router";

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
        "relative px-3 py-1.5 text-sm font-medium transition-all duration-150",
        isActive
          ? "text-brand-700 dark:text-brand-400"
          : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white",
      )}
    >
      {name}
      {isActive && (
        <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-brand-500 dark:bg-brand-400" />
      )}
    </Link>
  );
}

interface LanguagePickerProps {
  languages: Array<{ name: string; code: string }>;
}

export function LanguagePicker({ languages }: LanguagePickerProps) {
  return (
    <nav className="flex items-center gap-3">
      <span className="text-sm text-gray-500 dark:text-gray-400">
        Language:
      </span>
      <div className="flex items-center rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
        <Language name="OpenTofu" code={null} />
        {languages.map(({ name, code }) => (
          <Language key={code} name={name} code={code} />
        ))}
      </div>
    </nav>
  );
}
