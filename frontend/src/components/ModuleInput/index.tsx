import { HeadingLink } from "../HeadingLink";
import { Markdown } from "../Markdown";
import { Paragraph } from "../Paragraph";

interface ModuleInputProps {
  name: string;
  type: string;
  description: string;
  defaultValue?: unknown;
}

export function ModuleInput({
  name,
  type,
  description,
  defaultValue,
}: ModuleInputProps) {
  return (
    <li>
      <h4 id={name} className="group scroll-mt-5 font-bold">
        {name}{" "}
        <code className="text-mono text-sm text-purple-700 dark:text-purple-300">
          ({type})
        </code>
        <HeadingLink id={name} label={`${name} input`} />
      </h4>
      <Paragraph className="mt-1 ml-4"><Markdown text={description} /></Paragraph>
      {!!defaultValue && (
        <Paragraph className="mt-2 ml-4">
          Default value:{" "}
          <code className="text-mono break-words text-sm text-purple-700 dark:text-purple-300">
            {JSON.stringify(defaultValue)}
          </code>
        </Paragraph>
      )}
    </li>
  );
}
