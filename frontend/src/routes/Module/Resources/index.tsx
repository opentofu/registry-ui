import { Paragraph } from "@/components/Paragraph";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getModuleVersionDataQuery } from "../query";
import { EmptyState } from "../components/EmptyState";
import { useModuleParams } from "../hooks/useModuleParams";

interface ResourceProps {
  address: string;
}

function Resource({ address }: ResourceProps) {
  return (
    <li className="text-mono text-purple-700 dark:text-purple-300">
      {address}
    </li>
  );
}

export function ModuleResources() {
  const { namespace, name, target, version } = useModuleParams();

  const { data } = useSuspenseQuery(
    getModuleVersionDataQuery(namespace, name, target, version),
  );

  return (
    <div className="p-5">
      <h3 className="mb-2 text-3xl font-semibold">Resources</h3>
      <Paragraph>
        When using this module, it may create some resources. Below is a list of
        the resources that the module may create. These resources are
        identified by their unique address. However it is possible that some of
        these resources may either be not created at all, or multiples of them
        may be be created.
      </Paragraph>

      {data.resources.length === 0 && (
        <EmptyState
          text="This module does not have any outputs."
          className="mt-5"
        />
      )}

      {data.resources.length > 0 && (
        <ul className="mt-6 space-y-5">
          {data.resources.map((resource) => (
            <Resource key={resource.address} address={resource.address} />
          ))}
        </ul>
      )}
    </div>
  );
}
