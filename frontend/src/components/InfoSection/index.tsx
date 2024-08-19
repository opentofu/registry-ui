import { ReactNode } from "react";

interface InfoSectionProps {
  children: ReactNode;
}

export function InfoSection({ children }: InfoSectionProps) {
  return (
    <div className="bg-gray-100 p-4 dark:bg-blue-900">
      <dl className="flex flex-col gap-2">{children}</dl>
    </div>
  );
}

interface InfoSectionItemProps {
  label: string;
  children: ReactNode;
}

export function InfoSectionItem({ label, children }: InfoSectionItemProps) {
  return (
    <div className="flex items-center gap-2">
      <dt className="w-48 text-gray-700 dark:text-gray-300">{label}</dt>
      <dd>{children}</dd>
    </div>
  );
}
