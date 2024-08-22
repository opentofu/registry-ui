import { NotFoundPageError } from "@/utils/errors";
import { Navigate, useLocation, useRouteError } from "react-router-dom";
import { useProviderParams } from "../hooks/useProviderParams";

export function ProviderDocsError() {
  const routeError = useRouteError() as Error;
  const location = useLocation();
  const { namespace, provider, version } = useProviderParams();

  if (routeError instanceof NotFoundPageError && location.state?.fromVersion) {
    return (
      <Navigate to={`/provider/${namespace}/${provider}/${version}`} replace />
    );
  }

  throw routeError;
}
