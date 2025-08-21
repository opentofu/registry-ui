import { Link } from "react-router";
import { info } from "../../icons/info";
import { Icon } from "../Icon";

interface OldVersionBannerProps {
  latestVersionLink: string;
}

export function OldVersionBanner({ latestVersionLink }: OldVersionBannerProps) {
  return (
    <div className="flex items-center gap-4 bg-blue-200 p-4 dark:bg-blue-850">
      <Icon path={info} className="size-4 text-blue-700 dark:text-blue-500" />
      <span className="text-blue-900 dark:text-white">
        You are viewing a previous version.{" "}
        <Link to={latestVersionLink} className="underline">
          View the latest version
        </Link>
      </span>
    </div>
  );
}
