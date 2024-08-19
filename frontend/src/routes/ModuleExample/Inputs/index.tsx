import { EmptyState } from "@/routes/Module/components/EmptyState";

export function ModuleExampleInputs() {
  return (
    <div className="p-5">
      <h3 className="mb-6 text-3xl font-semibold">Inputs</h3>

      <EmptyState text="This module does not have any inputs." />
    </div>
  );
}
