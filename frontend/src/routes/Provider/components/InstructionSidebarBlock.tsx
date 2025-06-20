import { Code } from "@/components/Code";
import { Paragraph } from "@/components/Paragraph";
import { SidebarBlock } from "@/components/SidebarBlock";
import { useSuspenseQuery } from "@tanstack/react-query";
import { ReactNode } from "react";
import { useProviderParams } from "../hooks/useProviderParams";
import { getProviderDataQuery } from "../query";

function Block({ children }: { children: ReactNode }) {
  return (
    <SidebarBlock title="How to use this provider">
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

export function ProviderInstructionSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(getProviderDataQuery(namespace, provider));

  // strip the v from the version to display in the provider instructions
  // e.g. v0.1.0 -> 0.1.0
  const versionConstraint = version.replace(/^v/, "");

  const instruction = `terraform {
  required_providers {
    ${data.addr.name} = {
      source = "${data.addr.namespace}/${data.addr.name}"
      version = "${versionConstraint}"
    }
  }
}

provider "${data.addr.name}" {
  # Configuration options
}`;

  return (
    <Block>
      <Code value={instruction} language="hcl" />
    </Block>
  );
}

export function ProviderInstructionSidebarBlockSkeleton() {
  return (
    <Block>
      <span className="flex h-72 w-full animate-pulse bg-gray-500/25" />
    </Block>
  );
}
