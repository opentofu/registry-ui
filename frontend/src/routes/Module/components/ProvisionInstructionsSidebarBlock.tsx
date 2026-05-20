import { Code } from "@/components/Code";
import { Paragraph } from "@/components/Paragraph";
import { SidebarBlock } from "@/components/SidebarBlock";
import { getModuleDataQuery } from "@/routes/Module/query";
import { useSuspenseQuery } from "@tanstack/react-query";
import { ReactNode } from "react";
import { useModuleParams } from "../hooks/useModuleParams";

function Block({ children }: { children: ReactNode }) {
  return (
    <SidebarBlock title="How to use this module">
      <Paragraph className="my-3 text-sm text-gray-600 dark:text-gray-400">
        Add this to your configuration and run{" "}
        <code className="px-1 py-0.5 rounded bg-gray-200 dark:bg-gray-800 text-gray-800 dark:text-gray-200 font-mono text-xs">
          tofu init
        </code>
      </Paragraph>
      {children}
    </SidebarBlock>
  );
}

export function ModuleProvisionInstructionsSidebarBlock() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleDataQuery(namespace, name, target),
  );

  const instruction = `module "${data.addr.name}" {
  source = "${data.addr.namespace}/${data.addr.name}/${data.addr.target}"
  version = "${version}"
}`;

  return (
    <Block>
      <Code value={instruction} language="hcl" />
    </Block>
  );
}

export function ModuleProvisionInstructionsSidebarBlockSkeleton() {
  return (
    <Block>
      <span className="flex h-32 w-full animate-pulse bg-gray-500/25" />
    </Block>
  );
}
