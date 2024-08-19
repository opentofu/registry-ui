import { info } from "../../icons/info";
import { Icon } from "../Icon";

export function OldVersionBanner() {
  return (
    <div className="flex items-center gap-4 bg-blue-200 p-4 dark:bg-blue-850">
      <Icon path={info} className="size-4 text-blue-700 dark:text-blue-500" />
      <span className="text-blue-900 dark:text-white">
         You are not viewing the latest version. Use the sidebar on the right to select the latest version.
      </span>
    </div>
  );
}
