import { SidebarBlock } from "@/components/SidebarBlock";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getModuleVersionDataQuery } from "../query";
import { useState } from "react";
import { useModuleParams } from "../hooks/useModuleParams";
import { Paragraph } from "@/components/Paragraph";

export function ModuleSubmodulesSidebarBlock() {
  const [expanded, setExpanded] = useState(false);

  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const submodules = Object.keys(data.submodules);

  const visibleSubmodules = expanded ? submodules : submodules.slice(0, 5);

  return (
    <SidebarBlock title="Submodules">
      {submodules.length === 0 && (
        <Paragraph>This module does not have any submodules.</Paragraph>
      )}

      {submodules.length > 0 && (
        <ul className="mt-4 flex flex-col gap-4">
          {visibleSubmodules.map((submodule) => (
            <li key={submodule}>
              <Link
                to={`submodule/${submodule}`}
                className="text-inherit underline underline-offset-2"
              >
                {submodule}
              </Link>
            </li>
          ))}

          {submodules.length > 5 && (
            <li>
              <button
                type="button"
                onClick={() => setExpanded(!expanded)}
                className="text-gray-700 underline underline-offset-2 dark:text-gray-300"
              >
                {expanded ? "Show less" : "Show all submodules"}
              </button>
            </li>
          )}
        </ul>
      )}
    </SidebarBlock>
  );
}
