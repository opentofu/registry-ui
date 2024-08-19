import { HeadingLink } from "@/components/HeadingLink";
import { Paragraph } from "@/components/Paragraph";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useParams } from "react-router-dom";
import { definitions } from "@/api";
import { EmptyState } from "../components/EmptyState";
interface InputProps {
  name: string;
  type: string;
  description: string;
  defaultValue?: unknown;
}

function Input({ name, type, description, defaultValue }: InputProps) {
  return (
    <li>
      <h4 id={name} className="group scroll-mt-24">
        {name}{" "}
        <code className="text-mono text-sm text-purple-700 dark:text-purple-300">
          ({type})
        </code>
        <HeadingLink id={name} label={`${name} input`} />
      </h4>
      <Paragraph className="mt-1">{description}</Paragraph>
      {!!defaultValue && (
        <Paragraph className="mt-2">
          Default value:{" "}
          <code className="text-mono break-words text-sm text-purple-700 dark:text-purple-300">
            {JSON.stringify(defaultValue)}
          </code>
        </Paragraph>
      )}
    </li>
  );
}

export function ModuleInputs() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const requiredInputs: Array<definitions["Variable"] & { name: string }> = [];
  const optionalInputs: Array<definitions["Variable"] & { name: string }> = [];

  for (const [name, input] of Object.entries(data.variables)) {
    const result = { name, ...input };

    if (input.required) {
      requiredInputs.push(result);
    } else {
      optionalInputs.push(result);
    }
  }

  return (
    <>
      <div className="border-b border-gray-200 p-5 dark:border-gray-800">
        <h3 className="mb-2 text-3xl font-semibold">Required inputs</h3>
        <Paragraph>
          Because these variables do not have a default value defined by the
          module author they must be specified in the{" "}
          <code className="text-mono text-sm text-purple-700 dark:text-purple-300">
            module
          </code>{" "}
          block when using this module.
        </Paragraph>

        {requiredInputs.length === 0 && (
          <EmptyState
            text="This module does not have any required inputs."
            className="mt-5"
          />
        )}

        {requiredInputs.length > 0 && (
          <ul className="mt-6 space-y-5">
            {requiredInputs.map((input) => (
              <Input
                key={input.name}
                name={input.name}
                type={input.type}
                description={input.description}
              />
            ))}
          </ul>
        )}
      </div>
      <div className="p-5">
        <h3 className="mb-2 text-3xl font-semibold">Optional inputs</h3>
        <Paragraph>
          These inputs are optional and do not need to be specified in the{" "}
          <code className="text-mono text-sm text-purple-700 dark:text-purple-300">
            module
          </code>{" "}
          block when utilizing this module because they come with default values
          defined by the module author. However, to override the default value,
          you can specify these variables in the{" "}
          <code className="text-mono text-sm text-purple-700 dark:text-purple-300">
            module
          </code>{" "}
          block.
        </Paragraph>

        {optionalInputs.length === 0 && (
          <EmptyState
            text="This module does not have any optional inputs."
            className="mt-5"
          />
        )}

        {optionalInputs.length > 0 && (
          <ul className="mt-6 space-y-5">
            {optionalInputs.map((input) => (
              <Input
                key={input.name}
                name={input.name}
                type={input.type}
                description={input.description}
                defaultValue={input.default}
              />
            ))}
          </ul>
        )}
      </div>
    </>
  );
}
