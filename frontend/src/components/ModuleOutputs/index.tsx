import { EmptyState } from "@/components/EmptyState";
import { definitions } from "@/api";
import { ModuleOutput } from "../ModuleOutput";

interface ModuleOutputsProps {
  outputs: Record<string, definitions["Output"]>;
}

export function ModuleOutputs({ outputs }: ModuleOutputsProps) {
  const outputsWithNames = Object.entries(outputs).map(([name, output]) => ({
    name,
    ...output,
  }));

  return (
    <div className="p-5">
      <h3 className="mb-6 text-3xl font-semibold">Outputs</h3>

      {outputsWithNames.length === 0 && (
        <EmptyState text="This module does not have any outputs." />
      )}

      {outputsWithNames.length > 0 && (
        <ul className="mt-6 space-y-5">
          {outputsWithNames.map((output) => (
            <ModuleOutput
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
