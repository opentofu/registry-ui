import { SidebarLayout } from "@/components/SidebarLayout";
import { useLoaderData, useMatches } from "react-router";
import { SidebarPanel } from "@/components/SidebarPanel";
import { DocsSidebarMenu } from "./components/SidebarMenu";
import { MetaTags } from "@/components/MetaTags";
import { Document } from "./types";

import { useLocation } from "react-router";

export function Docs() {
  const matches = useMatches();
  const location = useLocation();
  
  // Find the deepest match that has loader data
  const activeMatch = matches
    .filter(match => match.data)
    .reverse()[0];
  
  const docs = activeMatch?.data as Document;

  return (
    <SidebarLayout
      key={location.pathname}
      before={
        <SidebarPanel>
          <DocsSidebarMenu />
        </SidebarPanel>
      }
      showBreadcrumbs={true}
    >
      <MetaTags title={docs?.data?.title} description={docs?.data?.description} />
      <div
        className="p-5"
        dangerouslySetInnerHTML={{ __html: docs?.content || "" }}
      />
    </SidebarLayout>
  );
}
