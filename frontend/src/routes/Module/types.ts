export interface ModuleRouteContext {
  version: string;
  rawVersion: string;
  namespace: string | undefined;
  name: string | undefined;
  target: string | undefined;
}
