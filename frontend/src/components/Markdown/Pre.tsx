import { HTMLAttributes, ReactElement } from "react";
import { Code } from "../Code";

export function MarkdownPre({ children }: HTMLAttributes<HTMLPreElement>) {
  if (!children) {
    return null;
  }

  const child = children as ReactElement;

  if (!child.props) {
    return (
      <pre className="overflow-auto bg-gray-200 p-4 text-sm dark:bg-blue-950">
        {children}
      </pre>
    );
  }

  if (!child.props.children) {
    return null;
  }

  const language = child.props.className?.match(/language-(\w+)/)?.[1];

  return (
    <Code
      value={child.props.children as string}
      language={language || "plaintext"}
      className="mt-5"
    />
  );
}
