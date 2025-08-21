import { Fragment, useState } from "react";
import { definitions } from "@/api";
import { Icon } from "../Icon";
import { info } from "@/icons/info";
import { groupBy } from "es-toolkit";

interface LicenseInfoProps {
  license?: definitions["LicenseList"];
}

export function LicenseInfo({ license }: LicenseInfoProps) {
  const [showPopover, setShowPopover] = useState(false);

  if (license === undefined) {
    return <span className="h-em flex w-24 animate-pulse bg-gray-500/25" />;
  }

  if (license === null || license.length === 0) {
    return (
      <div className="flex items-center gap-2">
        <span className="text-base font-medium text-gray-900 dark:text-white">
          None detected
        </span>
        <Icon
          title="The OpenTofu Search indexing couldn't detect a license with enough confidence to display this content."
          path={info}
          className="size-3.5 text-gray-500 dark:text-gray-400"
        />
      </div>
    );
  }

  const sortedLicenses = [...license].sort(
    (a, b) => b.confidence - a.confidence,
  );
  const groupedLicenses = groupBy(sortedLicenses, (license) => license.link);
  const licenseEntries = Object.entries(groupedLicenses);

  // Get the main license (highest confidence)
  const mainLicense = licenseEntries[0];
  const hasMultipleLicenses = licenseEntries.length > 1;

  const mainLicenseName = mainLicense[1]
    .map((license) => license.spdx)
    .join(", ");

  return (
    <div className="flex items-center gap-2">
      <a
        href={mainLicense[0]}
        target="_blank"
        rel="noreferrer noopener"
        className="hover:text-brand-600 dark:hover:text-brand-500 text-base font-medium text-gray-900 transition-colors dark:text-white"
      >
        {mainLicenseName}
      </a>

      {hasMultipleLicenses && (
        <div className="relative">
          <button
            onMouseEnter={() => setShowPopover(true)}
            onMouseLeave={() => setShowPopover(false)}
            className="flex items-center justify-center"
          >
            <Icon
              title="Multiple licenses detected"
              path={info}
              className="size-3.5 text-gray-500 transition-colors hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
            />
          </button>

          {showPopover && (
            <div className="absolute bottom-full left-1/2 z-10 mb-2 -translate-x-1/2 transform">
              <div className="min-w-48 rounded-lg border border-gray-200 bg-white p-3 shadow-lg dark:border-gray-700 dark:bg-gray-800">
                <div className="mb-2 text-sm font-medium text-gray-900 dark:text-white">
                  All detected licenses:
                </div>
                <ul className="space-y-2">
                  {licenseEntries.map(([link, licenses]) => (
                    <li key={link} className="text-sm">
                      <a
                        href={link}
                        className="block text-blue-600 transition-colors hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                        target="_blank"
                        rel="noreferrer noopener"
                      >
                        {licenses[0].file}
                      </a>
                      <div className="text-gray-600 dark:text-gray-400">
                        {licenses.map((license, index, arr) => (
                          <Fragment key={license.spdx}>
                            {license.spdx}
                            {index < arr.length - 1 && ", "}
                          </Fragment>
                        ))}
                      </div>
                    </li>
                  ))}
                </ul>
                <div className="mt-2 border-t border-gray-100 pt-2 dark:border-gray-700">
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    License detection is best effort. Please see the linked
                    license files for details.
                  </div>
                </div>
              </div>
              {/* Arrow */}
              <div className="absolute top-full left-1/2 -translate-x-1/2 transform">
                <div className="h-0 w-0 border-t-4 border-r-4 border-l-4 border-t-gray-200 border-r-transparent border-l-transparent dark:border-t-gray-700"></div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
