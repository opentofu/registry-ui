import { HeadingLink } from "@/components/HeadingLink";
import { Paragraph } from "@/components/Paragraph";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { getModuleVersionDataQuery } from "../query";
import { EmptyState } from "../components/EmptyState";

interface OutputProps {
  name: string;
  description: string;
}

function Output({ name, description }: OutputProps) {
  return (
    <li>
      <h4 id={name} className="group scroll-mt-24">
        {name}
        <HeadingLink id={name} label={`${name} output`} />
      </h4>
      <Paragraph className="mt-1">{description}</Paragraph>
    </li>
  );
}

export function ModuleOutputs() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const outputs = Object.entries(data.outputs).map(([name, output]) => ({
    name,
    ...output,
  }));

  return (
    <div className="p-5">
      <h3 className="mb-6 text-3xl font-semibold">Outputs</h3>

      {outputs.length === 0 && (
        <EmptyState text="This module does not have any outputs." />
      )}

      {outputs.length > 0 && (
        <ul className="mt-6 space-y-5">
          {outputs.map((output) => (
            <Output
              key={output.name}
              name={output.name}
              description={output.description}
            />
          ))}
        </ul>
      )}
    </div>
  );
}
