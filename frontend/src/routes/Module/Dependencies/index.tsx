import { Paragraph } from "@/components/Paragraph";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { getModuleVersionDataQuery } from "../query";
import { EmptyState } from "../components/EmptyState";

interface ProviderDependencyProps {
  provider: string;
  namespace: string;
  version: string;
}

function ProviderDependency({
  provider,
  namespace,
  version,
}: ProviderDependencyProps) {
  return (
    <li>
      {provider}{" "}
      <span className="text-gray-600 dark:text-gray-500">
        ({namespace}/{provider})
      </span>
      <code className="text-mono text-purple-700 dark:text-purple-300">
        {" "}
        {version}
      </code>
    </li>
  );
}

export function ModuleDependencies() {
  const { namespace, name, target, version } = useParams<{
    namespace: string;
    name: string;
    target: string;
    version: string;
  }>();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <>
      <div className="border-b border-gray-200 p-5 dark:border-gray-800">
        <h3 className="mb-2 text-3xl font-semibold">Module dependencies</h3>
        <Paragraph>
          Dependencies are external modules that the module references. Any
          module that is not located in the same repository is considered
          external.
        </Paragraph>
        <EmptyState
          text="This module has no external module dependencies."
          className="mt-5"
        />
      </div>
      <div className="px-5 pt-5">
        <h3 className="mb-2 text-3xl font-semibold">Provider dependencies</h3>
        <Paragraph>
          Terraform plugins, known as Providers, will be installed automatically
          when you run{" "}
          <code className="text-mono text-purple-700 dark:text-purple-300">
            terraform init
          </code>
          , as long as they are accessible on the registry.
        </Paragraph>
        <ul className="mt-6 space-y-5">
          {data.dependencies.map((dependency) => (
            <ProviderDependency
              key={dependency.name}
              provider={dependency.name}
              namespace={""}
              version={dependency.version_constraint}
            />
          ))}
        </ul>
      </div>
    </>
  );
}
