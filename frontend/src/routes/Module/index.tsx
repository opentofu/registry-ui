import { Outlet } from "react-router";
import { UnifiedLayout } from "@/components/UnifiedLayout";
import { SidebarPanel } from "@/components/SidebarPanel";
import { Suspense } from "react";

import { ModuleHeader, ModuleHeaderSkeleton } from "./components/Header";
import { ModuleSideMenu, ModuleSideMenuSkeleton } from "./components/SideMenu";
import {
  ModuleProvisionInstructionsSidebarBlock,
  ModuleProvisionInstructionsSidebarBlockSkeleton,
} from "./components/ProvisionInstructionsSidebarBlock";
import { ModuleExamplesSidebarBlock } from "./components/ExamplesSidebarBlock";
import { ModuleSubmodulesSidebarBlock } from "./components/SubmodulesSidebarBlock";
import {
  ModuleVersionInfo,
  ModuleVersionInfoSkeleton,
} from "./components/VersionInfo";
import {
  ModuleVersionsSidebarBlock,
  ModuleVersionsSidebarBlockSkeleton,
} from "./components/VersionsSidebarBlock";

export function Module() {
  return (
    <UnifiedLayout
      sidebar={
        <SidebarPanel>
          <Suspense fallback={<ModuleSideMenuSkeleton />}>
            <ModuleSideMenu />
          </Suspense>
        </SidebarPanel>
      }
      afterSidebar={
        <SidebarPanel className="divide-y divide-gray-200 dark:divide-gray-800">
          <Suspense fallback={<ModuleVersionsSidebarBlockSkeleton />}>
            <ModuleVersionsSidebarBlock />
          </Suspense>

          <Suspense
            fallback={<ModuleProvisionInstructionsSidebarBlockSkeleton />}
          >
            <ModuleProvisionInstructionsSidebarBlock />
            <ModuleExamplesSidebarBlock />
            <ModuleSubmodulesSidebarBlock />
          </Suspense>
        </SidebarPanel>
      }
      useIDEHeader={true}
    >
      <div className="flex flex-col gap-5 px-5">
        <Suspense
          fallback={
            <>
              <ModuleHeaderSkeleton />
              <ModuleVersionInfoSkeleton />
            </>
          }
        >
          <ModuleHeader />
          <ModuleVersionInfo />
        </Suspense>
      </div>

      <Outlet />
    </UnifiedLayout>
  );
}
