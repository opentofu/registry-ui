import { Outlet } from "react-router-dom";
import { SidebarLayout } from "../../components/SidebarLayout";
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
    <SidebarLayout
      before={
        <SidebarPanel>
          <ModuleSubmoduleSideMenu />
        </SidebarPanel>
      }
      after={
        <SidebarPanel>
          <Suspense
            fallback={<ModuleSubmoduleProvisionInstructionsSidebarBlock />}
          >
            <ModuleSubmoduleProvisionInstructionsSidebarBlock />
          </Suspense>
        </SidebarPanel>
      }
    >
      <div className="px-5">
        <Suspense fallback={<ModuleSubmoduleHeaderSkeleton />}>
          <ModuleSubmoduleHeader />
        </Suspense>
      </div>

      <Outlet />
    </SidebarLayout>
  );
}
