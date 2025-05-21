import { Outlet } from "react-router";
import { SidebarLayout } from "../../components/SidebarLayout";
import { SidebarPanel } from "../../components/SidebarPanel";
import { Suspense } from "react";

import {
  ModuleExampleHeader,
  ModuleExampleHeaderSkeleton,
} from "./components/Header";
import { ModuleExampleSideMenu } from "./components/SideMenu";
import { ModuleExampleProvisionInstructionsSidebarBlock } from "./components/ProvisionInstructionsSidebarBlock";

export function ModuleExample() {
  return (
    <SidebarLayout
      before={
        <SidebarPanel>
          <ModuleExampleSideMenu />
        </SidebarPanel>
      }
      after={
        <SidebarPanel>
          <ModuleExampleProvisionInstructionsSidebarBlock />
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
