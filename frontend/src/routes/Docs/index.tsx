import { SidebarLayout } from "@/components/SidebarLayout";
import { useLoaderData } from "react-router-dom";
import { SidebarPanel } from "@/components/SidebarPanel";
import { DocsSidebarMenu } from "./components/SidebarMenu";
import { MetaTags } from "@/components/MetaTags";
import { Document } from "./types";

export function Docs() {
  const docs = useLoaderData() as Document;

  return (
    <SidebarLayout
      before={
        <SidebarPanel>
          <DocsSidebarMenu />
        </SidebarPanel>
      }
    >
      <MetaTags title={docs.data.title} description={docs.data.description} />
      <div className="p-5" dangerouslySetInnerHTML={{ __html: docs.content }} />
    </SidebarLayout>
  );
}
