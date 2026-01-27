import { HeadingLink } from "../HeadingLink";
import { Markdown } from "../Markdown";
import { Paragraph } from "../Paragraph";

interface ModuleInputProps {
  name: string;
  type: any;
  description: string;
  defaultValue?: unknown;
}

function formatType(type: any): string {
  if (typeof type === "string") {
    return type;
  }

  if (Array.isArray(type) && type.length > 0) {
    const [kind, elementType] = type;

    switch (kind) {
      case "list":
        return `list(${formatType(elementType)})`;
      case "set":
        return `set(${formatType(elementType)})`;
      case "map":
        return `map(${formatType(elementType)})`;
      case "object":
        return "object";
      case "tuple":
        return "tuple";
      default:
        return String(kind);
    }
  }

  // Should never happen, we should have either a simple string or an array here
  // but just in case, we return unknown
  return "unknown";
}

export function ModuleInput({
  name,
  type,
  description,
  defaultValue,
}: ModuleInputProps) {
  const showDefaultValue = defaultValue !== undefined;
  const formattedType = formatType(type);
  return (
    <li>
      <h4 id={name} className="group scroll-mt-5 font-bold">
        {name}{" "}
        <code className="text-mono text-sm text-purple-700 dark:text-purple-300">
          ({formattedType})
        </code>
        <HeadingLink id={name} label={`${name} input`} />
      </h4>
      <Paragraph className="mt-1 ml-4">
        <Markdown text={description} />
      </Paragraph>
      {showDefaultValue && (
        <Paragraph className="mt-2 ml-4">
          Default value:{" "}
          <code className="text-mono text-sm break-words text-purple-700 dark:text-purple-300">
            {JSON.stringify(defaultValue)}
          </code>
        </Paragraph>
      )}
    </li>
  );
}
