import { ReactNode } from "react";

interface SidebarBlockProps {
  title: ReactNode;
  children: ReactNode;
}

export function SidebarBlock({ title, children }: SidebarBlockProps) {
  return (
    <div className="px-4 py-4">
      <h4 className="mb-3 flex items-center gap-2 font-sans text-base font-semibold leading-none text-gray-900 dark:text-white">
        {title}
      </h4>

      {children}
    </div>
  );
}
