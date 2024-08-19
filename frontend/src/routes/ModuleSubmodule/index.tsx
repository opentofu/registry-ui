import {
  Await,
  Link,
  NavLink,
  Outlet,
  useRouteLoaderData,
} from "react-router-dom";
import { SidebarLayout } from "../../components/SidebarLayout";
import { SidebarPanel } from "../../components/SidebarPanel";
import { Suspense } from "react";

import { Breadcrumbs, BreadcrumbsSkeleton } from "../../components/Breadcrumbs";
import { TreeView } from "../../components/TreeView";

import { SidebarBlock } from "../../components/SidebarBlock";
import { Paragraph } from "../../components/Paragraph";
import { Code } from "../../components/Code";

import { PageTitle } from "../../components/PageTitle";

import { ModuleTabLink } from "./TabLink";
import { Icon } from "../../components/Icon";
import { cross } from "../../icons/cross";
import { arrow } from "../../icons/arrow";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import { chevron } from "../../icons/chevron";
import clsx from "clsx";

const instruction = `module "iam" {
  source = "terraform-aws-modules/iam/aws"
  version = "5.39.1"
}`;

export function ModuleSubmodule() {
  const { data, namespace, module, submodule } =
    useRouteLoaderData("module-submodule");

  return (
    <SidebarLayout
      before={
        <SidebarPanel>
          <Menu>
            <MenuButton className="relative mr-4 mt-4 bg-gray-200 px-4 py-2 text-left dark:bg-blue-850">
              <span className="text-gray-700 dark:text-gray-300">
                Submodule
              </span>
              <span className="flex text-inherit">{submodule}</span>
              <Icon
                path={chevron}
                className="absolute right-4 top-6 size-4 rotate-90 text-inherit"
              />
            </MenuButton>
            <MenuItems
              anchor="bottom start"
              className="mt-1 flex min-w-96 flex-col bg-gray-200 dark:bg-blue-850"
            >
              <MenuItem>
                <NavLink
                  to={`../iam-account`}
                  relative="path"
                  className={({ isActive }) =>
                    clsx(
                      "px-2 py-1 text-inherit data-[focus]:bg-black/10 dark:data-[focus]:bg-white/10",
                      isActive && "bg-black/15 dark:bg-white/15",
                    )
                  }
                >
                  iam-account
                </NavLink>
              </MenuItem>
              <MenuItem>
                <NavLink
                  to={`../iam-assumable-role-with-oidc`}
                  relative="path"
                  className={({ isActive }) =>
                    clsx(
                      "px-2 py-1 text-inherit data-[focus]:bg-black/10 dark:data-[focus]:bg-white/10",
                      isActive && "bg-black/15 dark:bg-white/15",
                    )
                  }
                >
                  iam-assumable-role-with-oidc
                </NavLink>
              </MenuItem>
              <MenuItem>
                <NavLink
                  to={`../iam-assumable-role-with-saml`}
                  relative="path"
                  className={({ isActive }) =>
                    clsx(
                      "px-2 py-1 text-inherit data-[focus]:bg-black/10 dark:data-[focus]:bg-white/10",
                      isActive && "bg-black/15 dark:bg-white/15",
                    )
                  }
                >
                  iam-assumable-role-with-saml
                </NavLink>
              </MenuItem>
              <MenuItem>
                <NavLink
                  to={`../iam-assumable-roles`}
                  relative="path"
                  className={({ isActive }) =>
                    clsx(
                      "px-2 py-1 text-inherit data-[focus]:bg-black/10 dark:data-[focus]:bg-white/10",
                      isActive && "bg-black/15 dark:bg-white/15",
                    )
                  }
                >
                  iam-assumable-roles
                </NavLink>
              </MenuItem>
            </MenuItems>
          </Menu>

          <TreeView className="mr-4 mt-2">
            <ModuleTabLink to="." end>
              Readme
            </ModuleTabLink>
            <ModuleTabLink to="inputs">Inputs (0)</ModuleTabLink>
            <ModuleTabLink to="outputs">Outputs (0)</ModuleTabLink>
            <ModuleTabLink to="dependencies">Dependencies (0)</ModuleTabLink>
            <ModuleTabLink to="resources">Resources (0)</ModuleTabLink>
          </TreeView>
        </SidebarPanel>
      }
      after={
        <SidebarPanel className="divide-gray-200 dark:divide-gray-800">
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
      <div className="mx-5">
        <Suspense fallback={<BreadcrumbsSkeleton />}>
          <Await resolve={data}>
            <Breadcrumbs />
          </Await>
        </Suspense>
      </div>

      <div className="border-b border-gray-200 px-5 pb-5 dark:border-gray-800">
        <header className="flex items-center gap-4">
          <Link
            to="../.."
            relative="path"
            className="flex size-8 items-center justify-center bg-gray-100 dark:bg-blue-900"
          >
            <Icon path={arrow} className="size-4 text-inherit" />
          </Link>

          <PageTitle>{submodule}</PageTitle>
        </header>
      </div>
      <div className="px-4 pt-4">
        <Outlet />
      </div>
    </SidebarLayout>
  );
}
