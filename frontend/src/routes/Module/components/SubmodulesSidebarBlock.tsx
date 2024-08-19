import { SidebarBlock } from "@/components/SidebarBlock";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { getModuleVersionDataQuery } from "../query";

export function ModuleSubmodulesSidebarBlock() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const submodules = Object.keys(data.submodules);

  if (submodules.length === 0) {
    return null;
  }

  return (
    <SidebarBlock title="Submodules">
      <ul className="mt-4 flex flex-col gap-4">
        {submodules.map((submodule) => (
          <li key={submodule}>
            <Link
              to={`submodule/${submodule}`}
              className="text-inherit underline underline-offset-2"
            >
              {submodule}
            </Link>
          </li>
        ))}

        <li>
          <button
            type="button"
            className="text-gray-700 underline underline-offset-2 dark:text-gray-300"
          >
            Show all submodules
          </button>
        </li>
      </ul>
    </SidebarBlock>
  );
}
