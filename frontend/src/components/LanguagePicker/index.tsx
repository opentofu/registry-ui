import { clsx } from "clsx";

interface LanguageProps {
  name: string;
  selected: boolean;
  onClick: () => void;
}

function Language({ name, selected, onClick }: LanguageProps) {
  return (
    <button
      onClick={onClick}
      className={clsx(
        "ml-2 border px-4 py-2 font-semibold text-inherit",
        selected
          ? "border-brand-500 bg-brand-500 dark:border-brand-800 dark:bg-brand-800 dark:text-brand-600"
          : "border-gray-200 dark:border-gray-800",
      )}
    >
      {name}
    </button>
  );
}

interface LanguagePickerProps {
  languages: Array<{ name: string; code: string }>;
  selected: string | null;
  onChange: (language: string | null) => void;
}

export function LanguagePicker({
  languages,
  selected,
  onChange,
}: LanguagePickerProps) {
  return (
    <nav className="flex items-center">
      <span className="mr-2 text-gray-700 dark:text-gray-300">
        Provider language
      </span>
      <Language
        name="OpenTofu"
        selected={!selected}
        onClick={() => onChange(null)}
      />
      {languages.map(({ name, code }) => (
        <Language
          key={code}
          name={name}
          selected={selected === code}
          onClick={() => onChange(code)}
        />
      ))}
    </nav>
  );
}
