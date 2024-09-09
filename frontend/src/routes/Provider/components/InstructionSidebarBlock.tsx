import { ReactNode } from "react";
import { Code } from "@/components/Code";
import { Paragraph } from "@/components/Paragraph";
import { SidebarBlock } from "@/components/SidebarBlock";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getProviderDataQuery } from "../query";
import { useProviderParams } from "../hooks/useProviderParams";

function Block({ children }: { children: ReactNode }) {
  return (
    <SidebarBlock title="How to use this provider">
      <Paragraph className="my-4">
        Copy this code into your OpenTofu configuration and run run{" "}
        <code className="text-sm text-purple-700 dark:text-purple-300">
          tofu init
        </code>{" "}
        to install this provider.
      </Paragraph>
      {children}
    </SidebarBlock>
  );
}

export function ProviderInstructionSidebarBlock() {
  const { namespace, provider, version } = useProviderParams();

  const { data } = useSuspenseQuery(getProviderDataQuery(namespace, provider));

  const instruction = `terraform {
  required_providers {
    ${data.addr.namespace} = {
      source = "${data.addr.namespace}/${data.addr.name}"
      version = "${version}"
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
