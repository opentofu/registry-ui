import { Outlet } from "react-router";
import { UnifiedLayout } from "../../components/UnifiedLayout";
import { SidebarPanel } from "../../components/SidebarPanel";
import { Suspense } from "react";

import {
  ModuleSubmoduleHeader,
  ModuleSubmoduleHeaderSkeleton,
} from "./components/Header";
import { ModuleSubmoduleSideMenu } from "./components/SideMenu";
import { ModuleSubmoduleProvisionInstructionsSidebarBlock } from "./components/ProvisionInstructionsSidebarBlock";

export function ModuleSubmodule() {
  return (
    <UnifiedLayout
      sidebar={
        <SidebarPanel>
          <ModuleSubmoduleSideMenu />
        </SidebarPanel>
      }
      afterSidebar={
        <SidebarPanel>
          <Suspense
            fallback={<ModuleSubmoduleProvisionInstructionsSidebarBlock />}
          >
            <ModuleSubmoduleProvisionInstructionsSidebarBlock />
          </Suspense>
        </SidebarPanel>
      }
      useIDEHeader={true}
    >
      <div className="px-5">
        <Suspense fallback={<ModuleSubmoduleHeaderSkeleton />}>
          <ModuleSubmoduleHeader />
        </Suspense>
      </div>

      <Outlet />
    </UnifiedLayout>
  );
}
