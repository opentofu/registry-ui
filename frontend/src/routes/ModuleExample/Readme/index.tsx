import { useSuspenseQuery } from "@tanstack/react-query";

import { Markdown } from "@/components/Markdown";
import { getModuleExampleReadmeQuery } from "../query";
import { useParams } from "react-router-dom";

export function ModuleExampleReadme() {
  const { namespace, name, target, version, example } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
    example: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleExampleReadmeQuery(namespace, name, target, version, example),
  );

  return (
    <div className="p-5">
      <Markdown text={data} />
    </div>
  );
}
