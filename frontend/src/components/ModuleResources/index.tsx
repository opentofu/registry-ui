import { Paragraph } from "@/components/Paragraph";
import { EmptyState } from "@/components/EmptyState";
import { definitions } from "@/api";

interface ModuleResourcesProps {
  resources: Array<definitions["Resource"]>;
}

export function ModuleResources({ resources }: ModuleResourcesProps) {
  return (
    <div className="p-5">
      <h3 className="mb-2 text-3xl font-semibold">Resources</h3>
      <Paragraph>
        When using this module, it may create some resources. Below is a list of
        the resources that the module may create. These resources are identified
        by their unique address. However it is possible that some of these
        resources may either be not created at all, or multiples of them may be
        be created.
      </Paragraph>

      {resources.length === 0 && (
        <EmptyState
          text="This module does not have any resources."
          className="mt-5"
        />
      )}

      {resources.length > 0 && (
        <ul className="mt-6 space-y-5">
          {resources.map((resource) => (
            <li
              key={resource.address}
              className="text-mono text-purple-700 dark:text-purple-300"
            >
              {resource.address}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
