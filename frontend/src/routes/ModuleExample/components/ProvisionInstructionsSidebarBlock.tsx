import { Code } from "@/components/Code";
import { Paragraph } from "@/components/Paragraph";
import { SidebarBlock } from "@/components/SidebarBlock";
import { getModuleDataQuery } from "@/routes/Module/query";
import { useSuspenseQuery } from "@tanstack/react-query";
import { ReactNode } from "react";
import { useModuleExampleParams } from "../hooks/useModuleExampleParams";

function Block({ children }: { children: ReactNode }) {
  return (
    <SidebarBlock title="Provision instructions">
      <Paragraph className="my-4">TBA</Paragraph>
      {children}
    </SidebarBlock>
  );
}

export function ModuleExampleProvisionInstructionsSidebarBlock() {
  const { namespace, name, target, version, example } =
    useModuleExampleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  const instruction = `module "${data.addr.name}_${example}" {
  source = "${data.addr.namespace}/${data.addr.name}/${data.addr.target}//examples/${example}"
  version = "${version}"
}`;

  return (
    <Block>
      <Code value={instruction} language="hcl" />
    </Block>
  );
}

export function ModuleExampleProvisionInstructionsSidebarBlockSkeleton() {
  return (
    <Block>
      <span className="flex h-72 w-full animate-pulse bg-gray-500/25" />
    </Block>
  );
}
