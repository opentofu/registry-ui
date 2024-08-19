import { TreeView } from "@/components/TreeView";
import { ModuleTabLink } from "../TabLink";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { useParams } from "react-router-dom";

export function ModuleSideMenu() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const inputsCount = Object.keys(data.variables).length;
  const outputsCount = Object.keys(data.outputs).length;
  const dependenciesCount = data.dependencies.length;
  const resourcesCount = data.resources.length;

  return (
    <TreeView className="mr-4 mt-4">
      <ModuleTabLink to="." end>
        Readme
      </ModuleTabLink>
      <ModuleTabLink to="inputs">Inputs ({inputsCount})</ModuleTabLink>
      <ModuleTabLink to="outputs">Outputs ({outputsCount})</ModuleTabLink>
      <ModuleTabLink to="dependencies">
        Dependencies ({dependenciesCount})
      </ModuleTabLink>
      <ModuleTabLink to="resources">Resources ({resourcesCount})</ModuleTabLink>
    </TreeView>
  );
}

export function ModuleSideMenuSkeleton() {
  return <div>Loading</div>;
}
