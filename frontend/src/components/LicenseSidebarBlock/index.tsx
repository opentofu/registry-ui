import { SidebarBlock } from "../SidebarBlock";
import { Fragment, ReactNode } from "react";
import { definitions } from "@/api";
import { Icon } from "../Icon";
import { info } from "@/icons/info";

interface BlockProps {
  license?: definitions["LicenseList"];
}

function LicenseSidebarBlockTitle() {
  return (
    <>
      License
      <Icon
        title="License detection is best effort. Please see the linked license file for details."
        path={info}
        className="size-4 text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
      />
    </>
  );
}

export function LicenseSidebarBlock(props: BlockProps) {
  let content: ReactNode;

  if (props.license === undefined) {
    content = <span className="flex h-em w-24 animate-pulse bg-gray-500/25" />;
  } else if (props.license === null || props.license.length === 0) {
    content = "None detected";
  } else {
    const sortedLicenses = [...props.license].sort(
      (a, b) => b.confidence - a.confidence,
    );

    const groupedLicenses = Object.groupBy(
      sortedLicenses,
      (license) => license.link,
    );

    const licenses = Object.entries(groupedLicenses).map(([link, license]) => (
      <li className="flex flex-col items-start gap-2" key={link}>
        <a
          href={link}
          className="underline"
          target="_blank"
          rel="noreferrer noopener"
        >
          {license[0].file}
        </a>
        <span className="text-gray-800 dark:text-gray-300">
          {license?.map((license, index, arr) => (
            <Fragment key={license.spdx}>
              {license.spdx}
              {index < arr.length - 1 && ", "}
            </Fragment>
          ))}
        </span>
      </li>
    ));

    content = <ul className="flex flex-col gap-4">{licenses}</ul>;
  }

  return (
    <SidebarBlock title={<LicenseSidebarBlockTitle />}>{content}</SidebarBlock>
  );
}

export function LicenceSidebarBlockSkeleton() {
  return <LicenseSidebarBlock />;
}
