export interface Crumb {
  to: string;
  label: string;
}

export function createCrumb(to: string, label: string): Crumb {
  return {
    to,
    label,
  };
}
