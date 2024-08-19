import { Highlight, Prism } from "prism-react-renderer";
import { theme } from "./theme";
import { suspend } from "suspend-react";

globalThis.Prism = Prism;

interface SyntaxHighlighterProps {
  value: string;
  language: string;
}

const languages: Record<string, () => Promise<void>> = {
  hcl: () => import("prismjs/components/prism-hcl"),
};

const aliases: Record<string, string> = {
  terraform: "hcl",
  tf: "hcl",
};

export function SyntaxHighlighter({ value, language }: SyntaxHighlighterProps) {
  if (aliases[language]) {
    language = aliases[language];
  }

  if (languages[language]) {
    suspend(languages[language], [language]);
  }

  const trimmedValue = value.replace(/^\n+|\n+$/g, "");

  return (
    <Highlight theme={theme} code={trimmedValue} language={language}>
      {({ tokens, getLineProps, getTokenProps }) => (
        <>
          {tokens.map((line, i) => (
            <div key={i} {...getLineProps({ line })}>
              {line.map((token, key) => (
                <span key={key} {...getTokenProps({ token })} />
              ))}
            </div>
          ))}
        </>
      )}
    </Highlight>
  );
}
