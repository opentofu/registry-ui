import { EmptyState } from "@/routes/Module/components/EmptyState";

export function ModuleExampleOutputs() {
  return (
    <div className="p-5">
      <h3 className="mb-6 text-3xl font-semibold">Outputs</h3>

      <EmptyState text="This module does not have any outputs." />
    </div>
  );
}
