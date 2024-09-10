import { NotFoundPageError } from "@/utils/errors";
import { Navigate, useLocation, useRouteError } from "react-router-dom";
import { useProviderParams } from "../hooks/useProviderParams";

function ProviderVersionRedirect() {
  const params = useProviderParams();

  return (
    <Navigate
      to={`/provider/${params.namespace}/${params.provider}/latest`}
      replace
    />
  );
}

export function ProviderError() {
  const routeError = useRouteError() as Error;
  const location = useLocation();

  if (routeError instanceof NotFoundPageError && location.state?.fromVersion) {
    return <ProviderVersionRedirect />;
  }

  throw routeError;
}
