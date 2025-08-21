import { TreeView } from "@/components/TreeView";
import { ModuleTabLink } from "../TabLink";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useModuleParams } from "../hooks/useModuleParams";

export function ModuleSideMenu() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const inputsCount = Object.keys(data.variables).length;
  const outputsCount = Object.keys(data.outputs).length;
  const dependenciesCount = data.dependencies.length;
  const resourcesCount = data.resources.length;

  const hasSchemaError = !!data.schema_error;

  return (
    <div className="p-4">
      <TreeView>
      <ModuleTabLink to="." end>
        Readme
      </ModuleTabLink>
      <ModuleTabLink to="inputs" count={inputsCount} disabled={hasSchemaError}>
        Inputs
      </ModuleTabLink>
      <ModuleTabLink
        to="outputs"
        count={outputsCount}
        disabled={hasSchemaError}
      >
        Outputs
      </ModuleTabLink>
      <ModuleTabLink
        to="dependencies"
        count={dependenciesCount}
        disabled={hasSchemaError}
      >
        Dependencies
      </ModuleTabLink>
      <ModuleTabLink
        to="resources"
        count={resourcesCount}
        disabled={hasSchemaError}
      >
        Resources
      </ModuleTabLink>
      </TreeView>
    </div>
  );
}

export function ModuleSideMenuSkeleton() {
  return (
    <div className="p-4 flex animate-pulse flex-col gap-5">
      <span className="flex h-em w-48 bg-gray-500/25" />
      <span className="flex h-em w-52 bg-gray-500/25" />
      <span className="flex h-em w-36 bg-gray-500/25" />
      <span className="flex h-em w-64 bg-gray-500/25" />
      <span className="flex h-em w-56 bg-gray-500/25" />
    </div>
  );
}
