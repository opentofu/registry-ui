import { defer, LoaderFunction } from "react-router-dom";
import { queryClient } from "../../query";
import { getNamespaceProvidersQuery, getProvidersQuery } from "./query";

export const providersLoader = () => {
  const index = queryClient.ensureQueryData(getProvidersQuery());
  return defer({ index });
};

export const namespaceProvidersLoader: LoaderFunction = ({ params }) => {
  const index = queryClient.ensureQueryData(
    getNamespaceProvidersQuery(params.namespace),
  );

  return {
    index,
    namespace: params.namespace,
  };
};
