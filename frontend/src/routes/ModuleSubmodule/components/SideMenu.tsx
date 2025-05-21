import { Icon } from "@/components/Icon";
import { TreeView } from "@/components/TreeView";
import { chevron } from "@/icons/chevron";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import clsx from "clsx";
import { NavLink } from "react-router";
import { ModuleTabLink } from "./TabLink";
import { useSuspenseQueries } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "@/routes/Module/query";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";
import { getModuleSubmoduleDataQuery } from "../query";

export function ModuleSubmoduleSideMenu() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const [{ data }, { data: submoduleData }] = useSuspenseQueries({
    queries: [
      getModuleVersionDataQuery(namespace, name, target, version),
      getModuleSubmoduleDataQuery(namespace, name, target, version, submodule),
    ],
  });

  const submodules = Object.keys(data.submodules);
  const inputsCount = Object.keys(submoduleData.variables).length;
  const outputsCount = Object.keys(submoduleData.outputs).length;
  const dependenciesCount = submoduleData.dependencies.length;
  const resourcesCount = submoduleData.resources.length;

  return (
    <>
      <Menu>
        <MenuButton className="relative mr-4 mt-4 bg-gray-200 px-4 py-2 text-left dark:bg-blue-850">
          <span className="text-gray-700 dark:text-gray-300">Submodule</span>
          <span className="flex text-inherit">{submodule}</span>
          <Icon
            path={chevron}
            className="absolute right-4 top-6 size-4 rotate-90 text-inherit"
          />
        </MenuButton>
        <MenuItems
          anchor="bottom start"
          className="mt-1 flex min-w-96 flex-col bg-gray-200 dark:bg-blue-850"
        >
          {submodules.map((submodule) => (
            <MenuItem key={submodule}>
              <NavLink
                to={`../${submodule}`}
                relative="path"
                className={({ isActive }) =>
                  clsx(
                    "px-2 py-1 text-inherit data-[focus]:bg-black/10 dark:data-[focus]:bg-white/10",
                    isActive && "bg-black/15 dark:bg-white/15",
                  )
                }
              >
                {submodule}
              </NavLink>
            </MenuItem>
          ))}
        </MenuItems>
      </Menu>
      <TreeView className="mr-4 mt-2">
        <ModuleTabLink to="." end>
          Readme
        </ModuleTabLink>
        <ModuleTabLink to="inputs">Inputs ({inputsCount})</ModuleTabLink>
        <ModuleTabLink to="outputs">Outputs ({outputsCount})</ModuleTabLink>
        <ModuleTabLink to="dependencies">
          Dependencies ({dependenciesCount})
        </ModuleTabLink>
        <ModuleTabLink to="resources">
          Resources ({resourcesCount})
        </ModuleTabLink>
      </TreeView>
    </>
  );
}
