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
      <Paragraph className="my-4">
        Copy this code into your OpenTofu configuration and run{" "}
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
