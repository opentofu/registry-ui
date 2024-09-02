import { Icon } from "@/components/Icon";
import { warning } from "@/icons/warning";

export function ModuleSchemaError() {
  return (
    <div className="flex items-center gap-3 bg-brand-150 px-4 py-4 dark:bg-brand-800">
      <Icon
        path={warning}
        className="size-em text-brand-700 dark:text-brand-600"
      />
      We were unable to parse the schema for this module.
    </div>
  );
}
