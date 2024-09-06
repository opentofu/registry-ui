import { SidebarLayout } from "@/components/SidebarLayout";
import { useLoaderData } from "react-router-dom";
import { Markdown } from "@/components/Markdown";
import { SidebarPanel } from "@/components/SidebarPanel";
import { DocsSidebarMenu } from "./components/SidebarMenu";

export function Docs() {
  const data = useLoaderData();

  return (
    <SidebarLayout
      before={
        <SidebarPanel>
          <DocsSidebarMenu />
        </SidebarPanel>
      }
    >
      <div className="p-5">
        <Markdown text={data.content} />
      </div>
    </SidebarLayout>
  );
}
