import { SidebarBlock } from "../SidebarBlock";
import { Fragment, ReactNode } from "react";
import { definitions } from "@/api";
import { Icon } from "../Icon";
import { info } from "@/icons/info";
import { groupBy } from "es-toolkit";

interface BlockProps {
  license?: definitions["LicenseList"];
}

function LicenseSidebarBlockTitle() {
  return (
    <div className="flex items-center gap-1.5">
      License
      <Icon
        title="License detection is best effort. Please see the linked license file for details."
        path={info}
        className="size-3.5 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
      />
    </div>
  );
}

export function LicenseSidebarBlock(props: BlockProps) {
  let content: ReactNode;

  if (props.license === undefined) {
    content = <span className="flex h-em w-24 animate-pulse bg-gray-500/25" />;
  } else if (props.license === null || props.license.length === 0) {
    content = (
      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        None detected
        <Icon
          title="The OpenTofu Search indexing couldn't detect a license with enough confidence to display this content."
          path={info}
          className="size-3.5"
        />
      </div>
    );
  } else {
    const sortedLicenses = [...props.license].sort(
      (a, b) => b.confidence - a.confidence,
    );

    const groupedLicenses = groupBy(sortedLicenses, (license) => license.link);

    const licenses = Object.entries(groupedLicenses).map(([link, license]) => (
      <li className="flex flex-col items-start gap-1" key={link}>
        <a
          href={link}
          className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
          target="_blank"
          rel="noreferrer noopener"
        >
          {license[0].file}
        </a>
        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
          {license?.map((license, index, arr) => (
            <Fragment key={license.spdx}>
              {license.spdx}
              {index < arr.length - 1 && ", "}
            </Fragment>
          ))}
        </span>
      </li>
    ));

    content = <ul className="flex flex-col gap-3">{licenses}</ul>;
  }

  return (
    <SidebarBlock title={<LicenseSidebarBlockTitle />}>{content}</SidebarBlock>
  );
}

export function LicenceSidebarBlockSkeleton() {
  return <LicenseSidebarBlock />;
}
