import { SidebarBlock } from "@/components/SidebarBlock";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { getModuleVersionDataQuery } from "../query";
import { useState } from "react";

export function ModuleExamplesSidebarBlock() {
  const [expanded, setExpanded] = useState(false);

  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  const examples = Object.keys(data.examples);

  if (examples.length === 0) {
    return null;
  }

  const visibleExamples = expanded ? examples : examples.slice(0, 5);

  return (
    <SidebarBlock title="Examples">
      <ul className="mt-4 flex flex-col gap-4">
        {visibleExamples.map((example) => (
          <li key={example}>
            <Link
              to={`example/${example}`}
              className="text-inherit underline underline-offset-2"
            >
              {example}
            </Link>
          </li>
        ))}

        {examples.length > 5 && (
          <li>
            <button
              type="button"
              onClick={() => setExpanded(!expanded)}
              className="text-gray-700 underline underline-offset-2 dark:text-gray-300"
            >
              {expanded ? "Show less" : "Show all examples"}
            </button>
          </li>
        )}
      </ul>
    </SidebarBlock>
  );
}
