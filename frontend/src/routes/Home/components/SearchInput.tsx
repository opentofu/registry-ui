import { Icon } from "@/components/Icon";
import { search as searchIcon } from "@/icons/search";

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent) => void;
  onFocus?: () => void;
  placeholder?: string;
  autoFocus?: boolean;
  showClearButton?: boolean;
  onClear?: () => void;
  size?: "small" | "large";
}

export function SearchInput({
  value,
  onChange,
  onKeyDown,
  onFocus,
  placeholder = "Search...",
  autoFocus = false,
  showClearButton = false,
  onClear,
  size = "large",
}: SearchInputProps) {
  const isSmall = size === "small";

  return (
    <div className="relative">
      <Icon
        path={searchIcon}
        className={`absolute top-1/2 -translate-y-1/2 text-gray-400 ${isSmall ? "left-3 size-4" : "left-4 size-5"
          }`}
      />
      <input
        type="text"
        aria-label="Search"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={onKeyDown}
        onFocus={onFocus}
        placeholder={placeholder}
        autoFocus={autoFocus}
        className={`focus:ring-brand-500 dark:focus:ring-brand-400 w-full rounded-xl border border-gray-200 bg-white text-gray-900 placeholder:text-gray-400 focus:border-transparent focus:ring-2 focus:outline-none dark:border-gray-700 dark:bg-blue-900 dark:text-gray-200 dark:placeholder-gray-400 ${isSmall
            ? "h-9 pr-9 pl-9 text-sm dark:border-gray-600 dark:bg-gray-800"
            : "h-14 pr-4 pl-12 text-base shadow-sm"
          } ${showClearButton && isSmall ? "pr-9" : ""} `}
      />
      {showClearButton && value && onClear && (
        <button
          aria-label="Clear search"
          onClick={onClear}
          className="absolute top-1/2 right-2.5 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        >
          <svg
            className="h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      )}
    </div>
  );
}
