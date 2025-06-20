import { Link } from "react-router";
import { PageTitle } from "@/components/PageTitle";
import { Breadcrumbs, BreadcrumbsSkeleton } from "@/components/Breadcrumbs";
import { ReactNode } from "react";
import { useSuspenseQuery } from "@tanstack/react-query";

import { getModuleDataQuery } from "@/routes/Module/query";
import { Icon } from "@/components/Icon";
import { arrow } from "@/icons/arrow";
import { useModuleSubmoduleParams } from "../hooks/useModuleSubmoduleParams";

interface WrapperProps {
  children: ReactNode;
}

function Wrapper({ children }: WrapperProps) {
  return (
    <div className="-mx-5 border-b border-gray-200 px-5 pb-5 dark:border-gray-800">
      {children}
    </div>
  );
}

export function ModuleSubmoduleHeader() {
  const { namespace, name, target, submodule } = useModuleSubmoduleParams();

  useSuspenseQuery(getModuleDataQuery(namespace, name, target));

  return (
    <Wrapper>
      <Breadcrumbs />
      <div className="flex items-center gap-4">
        <Link
          to="../.."
          relative="path"
          className="flex size-8 items-center justify-center bg-gray-100 dark:bg-blue-900"
        >
          <Icon path={arrow} className="size-4 text-inherit" />
        </Link>

        <PageTitle>{submodule}</PageTitle>
      </div>
    </Wrapper>
  );
}

export function ModuleSubmoduleHeaderSkeleton() {
  return (
    <Wrapper>
      <BreadcrumbsSkeleton />
    </Wrapper>
  );
}
