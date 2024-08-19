import { Outlet } from "react-router-dom";
import { SidebarLayout } from "../../components/SidebarLayout";
import { SidebarPanel } from "../../components/SidebarPanel";
import { Suspense } from "react";

import { SidebarBlock } from "../../components/SidebarBlock";
import { Paragraph } from "../../components/Paragraph";
import { Code } from "../../components/Code";

import {
  ModuleExampleHeader,
  ModuleExampleHeaderSkeleton,
} from "./components/Header";
import { ModuleExampleSideMenu } from "./components/SideMenu";

const instruction = `module "iam" {
  source = "terraform-aws-modules/iam/aws"
  version = "5.39.1"
}`;

export function ModuleExample() {
  return (
    <SidebarLayout
      before={
        <SidebarPanel>
          <ModuleExampleSideMenu />
        </SidebarPanel>
      }
      after={
        <SidebarPanel className="divide-y divide-gray-200 dark:divide-gray-800">
          <SidebarBlock title="Provision instructions">
            <Paragraph className="my-4">
              Copy this code info your OpenTofu configuration and add any variables necessary, then run{" "}
              <code className="text-sm text-purple-700 dark:text-purple-300">
                tofu init
              </code>
              .
            </Paragraph>
            <Code value={instruction} language="hcl" />
          </SidebarBlock>
        </SidebarPanel>
      }
    >
      <div className="px-5">
        <Suspense fallback={<ModuleExampleHeaderSkeleton />}>
          <ModuleExampleHeader />
        </Suspense>
      </div>

      <Outlet />
    </SidebarLayout>
  );
}
