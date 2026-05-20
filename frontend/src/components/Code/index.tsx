import { Suspense, useRef, useState } from "react";
import { SyntaxHighlighter } from "./SyntaxHighlighter";
import clsx from "clsx";
import { Icon } from "../Icon";
import { tick } from "../../icons/tick";
import { copy } from "../../icons/copy";

interface CodeProps {
  value: string;
  language: string;
  className?: string;
}

// TODO: make the button accessible
export function Code({ value, language, className }: CodeProps) {
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<number | null>(null);

  const copyToClipboard = async () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
    } finally {
      timeoutRef.current = setTimeout(() => setCopied(false), 1000);
    }
  };

  return (
    <div className={clsx("relative bg-gray-200 dark:bg-blue-950", className)}>
      <pre className="overflow-auto p-4 text-sm">
        <Suspense fallback={<code>{value}</code>}>
          <SyntaxHighlighter value={value} language={language} />
        </Suspense>
      </pre>
      <button
        className="absolute right-2 top-2 bg-gray-300 p-2 text-inherit dark:bg-blue-900"
        onClick={copyToClipboard}
        aria-label="Copy code to clipboard"
      >
        <Icon
          path={copied ? tick : copy}
          className={clsx(
            "size-4",
            copied && "text-green-600 dark:text-green-500",
          )}
        />
      </button>
    </div>
  );
}
