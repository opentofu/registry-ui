import { Outlet } from "react-router-dom";
import { SidebarLayout } from "../../components/SidebarLayout";
import { SidebarPanel } from "../../components/SidebarPanel";
import { Suspense } from "react";

import { ProviderHeader, ProviderHeaderSkeleton } from "./components/Header";
import {
  ProviderVersionInfo,
  ProviderVersionInfoSkeleton,
} from "./components/VersionInfo";
import {
  ProviderDocsMenu,
  ProviderDocsMenuSkeleton,
} from "./components/DocsMenu";
import {
  ProviderVersionsSidebarBlock,
  ProviderVersionsSidebarBlockSkeleton,
} from "./components/VersionsSidebarBlock";
import {
  ProviderInstructionSidebarBlock,
  ProviderInstructionSidebarBlockSkeleton,
} from "./components/InstructionSidebarBlock";
import {
  ProviderMetadataSidebarBlock,
  ProviderMetadataSidebarBlockSkeleton,
} from "./components/MetadataSidebarBlock";
import { ProviderPageTitle } from "./components/PageTitle";

export function Provider() {
  return (
    <SidebarLayout
      before={
        <SidebarPanel>
          <Suspense fallback={<ProviderDocsMenuSkeleton />}>
            <ProviderDocsMenu />
          </Suspense>
        </SidebarPanel>
      }
      after={
        <SidebarPanel className="divide-y divide-gray-200 dark:divide-gray-800">
          <Suspense
            fallback={
              <>
                <ProviderVersionsSidebarBlockSkeleton />
                <ProviderInstructionSidebarBlockSkeleton />
                <ProviderMetadataSidebarBlockSkeleton />
              </>
            }
          >
            <ProviderVersionsSidebarBlock />
            <ProviderInstructionSidebarBlock />
            <ProviderMetadataSidebarBlock />
          </Suspense>
        </SidebarPanel>
      }
    >
      <div className="px-5 pb-5">
        <Suspense
          fallback={
            <>
              <ProviderHeaderSkeleton />
              <ProviderVersionInfoSkeleton />
            </>
          }
        >
          <ProviderPageTitle />
          <ProviderHeader />
          <ProviderVersionInfo />
        </Suspense>

        <Outlet />
      </div>
    </SidebarLayout>
  );
}
