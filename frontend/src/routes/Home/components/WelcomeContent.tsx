import { Paragraph } from "@/components/Paragraph";
import { Icon } from "@/components/Icon";
import { info } from "@/icons/info";
import Logo from "@/logo.svg?react";

export function WelcomeContent() {
  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-1 flex-col items-center justify-center p-8 text-center">
        <Logo className="mb-8 w-32 h-32" />
        <h2 className="mb-4 max-w-2xl text-4xl leading-tight font-bold text-balance">
          Discover Infrastructure Building Blocks
        </h2>
        <Paragraph className="mb-8 max-w-2xl text-balance text-lg">
          Search through <strong>4,000+ providers</strong> and{" "}
          <strong>22,000+ modules</strong> to build and manage your
          infrastructure with OpenTofu
        </Paragraph>

        <div className="w-full max-w-2xl mt-8">
          <h3 className="mb-6 text-xl font-semibold text-gray-900 dark:text-gray-100">
            What can you find here?
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
            <div className="text-left">
              <h4 className="mb-3 text-lg font-medium text-gray-900 dark:text-gray-100">
                Providers
              </h4>
              <div className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                <p>• Cloud platforms (AWS, Azure, GCP)</p>
                <p>• Infrastructure services (Kubernetes, Docker)</p>
                <p>• Monitoring tools (Datadog, New Relic)</p>
                <p>• DevOps platforms (GitHub, GitLab)</p>
              </div>
            </div>

            <div className="text-left">
              <h4 className="mb-3 text-lg font-medium text-gray-900 dark:text-gray-100">
                Modules
              </h4>
              <div className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                <p>• VPC and networking setups</p>
                <p>• Security groups and IAM policies</p>
                <p>• Database and storage configurations</p>
                <p>• Complete application stacks</p>
              </div>
            </div>
          </div>

          <div
            role="alert"
            className="mt-5 rounded-lg border border-sky-200 bg-sky-50 px-4 py-3 text-left dark:border-sky-800 dark:bg-sky-950"
          >
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0">
                <Icon
                  path={info}
                  className="h-5 w-5 text-sky-600 dark:text-sky-400 mt-0.5"
                />
              </div>
              <div className="flex-1">
                <h4 className="mb-1 text-sm font-semibold text-sky-900 dark:text-sky-100">
                  Pro tip
                </h4>
                <p className="text-sm text-sky-800 dark:text-sky-200">
                  Try searching for "aws", "kubernetes", or "vpc" to discover
                  popular infrastructure patterns and components.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
