import { Code } from "@/components/Code";
import { Paragraph } from "@/components/Paragraph";
import { SidebarBlock } from "@/components/SidebarBlock";
import { getModuleDataQuery } from "@/routes/Module/query";
import { useSuspenseQuery } from "@tanstack/react-query";
import { ReactNode } from "react";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";

function Block({ children }: { children: ReactNode }) {
  return (
    <SidebarBlock title="How to use this module">
      <Paragraph className="my-4">
        Copy this code info your OpenTofu configuration and add any variables
        necessary, then run{" "}
        <code className="text-sm text-purple-700 dark:text-purple-300">
          tofu init
        </code>
        .
      </Paragraph>
      {children}
    </SidebarBlock>
  );
}

export function ModuleSubmoduleProvisionInstructionsSidebarBlock() {
  const { namespace, name, target, version, submodule } =
    useModuleSubmoduleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  const instruction = `module "${data.addr.name}_${submodule}" {
  source = "${data.addr.namespace}/${data.addr.name}/${data.addr.target}//submodules/${submodule}"
  version = "${version}"
}`;

  return (
    <Block>
      <Code value={instruction} language="hcl" />
    </Block>
  );
}

export function ModuleSubmoduleProvisionInstructionsSidebarBlockSkeleton() {
  return (
    <Block>
      <span className="flex h-72 w-full animate-pulse bg-gray-500/25" />
    </Block>
  );
}
