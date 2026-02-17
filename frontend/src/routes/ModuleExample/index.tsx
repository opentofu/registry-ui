import { Outlet } from "react-router";
import { UnifiedLayout } from "../../components/UnifiedLayout";
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
    <UnifiedLayout
      sidebar={
        <SidebarPanel>
          <ModuleExampleSideMenu />
        </SidebarPanel>
      }
      afterSidebar={
        <SidebarPanel>
          <ModuleExampleProvisionInstructionsSidebarBlock />
        </SidebarPanel>
      }
      useIDEHeader={true}
    >
      <div className="px-5">
        <Suspense fallback={<ModuleExampleHeaderSkeleton />}>
          <ModuleExampleHeader />
        </Suspense>
      </div>

      <Outlet />
    </UnifiedLayout>
  );
}
