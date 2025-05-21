import { Icon } from "@/components/Icon";
import { TreeView } from "@/components/TreeView";
import { chevron } from "@/icons/chevron";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import clsx from "clsx";
import { NavLink } from "react-router";
import { ModuleTabLink } from "./TabLink";
import { useSuspenseQueries } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "@/routes/Module/query";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";
import { getModuleExampleDataQuery } from "../query";

export function ModuleExampleSideMenu() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const [{ data }, { data: exampleData }] = useSuspenseQueries({
    queries: [
      getModuleVersionDataQuery(namespace, name, target, version),
      getModuleExampleDataQuery(namespace, name, target, version, example),
    ],
  });

  const examples = Object.keys(data.examples);
  const inputsCount = Object.keys(exampleData.variables).length;
  const outputsCount = Object.keys(exampleData.outputs).length;

  return (
    <>
      <Menu>
        <MenuButton className="relative mr-4 mt-4 bg-gray-200 px-4 py-2 text-left dark:bg-blue-850">
          <span className="text-gray-700 dark:text-gray-300">Example</span>
          <span className="flex text-inherit">{example}</span>
          <Icon
            path={chevron}
            className="absolute right-4 top-6 size-4 rotate-90 text-inherit"
          />
        </MenuButton>
        <MenuItems
          anchor="bottom start"
          className="mt-1 flex min-w-96 flex-col bg-gray-200 dark:bg-blue-850"
        >
          {examples.map((example) => (
            <MenuItem key={example}>
              <NavLink
                to={`../${example}`}
                relative="path"
                className={({ isActive }) =>
                  clsx(
                    "px-2 py-1 text-inherit data-[focus]:bg-black/10 dark:data-[focus]:bg-white/10",
                    isActive && "bg-black/15 dark:bg-white/15",
                  )
                }
              >
                {example}
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
      </TreeView>
    </>
  );
}
