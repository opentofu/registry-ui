export interface Crumb {
  to: string;
  label: string;
}

export function createCrumb(to: string, label: string | undefined = ""): Crumb {
  return {
    to,
    label,
  };
}
