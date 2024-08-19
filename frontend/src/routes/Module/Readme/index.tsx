import { useSuspenseQuery } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import { getModuleReadmeQuery } from "../query";
import { useParams } from "react-router-dom";

export function ModuleReadme() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleReadmeQuery(namespace, name, target, version),
  );

  return (
    <div className="p-5">
      <Markdown text={data} />
    </div>
  );
}
