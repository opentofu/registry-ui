import { ReactNode } from "react";

interface SidebarBlockProps {
  title: string;
  children: ReactNode;
}

export function SidebarBlock({ title, children }: SidebarBlockProps) {
  return (
    <div className="px-4 py-4">
      <h4 className="mb-4 font-sans text-xl font-semibold leading-none">
        {title}
      </h4>

      {children}
    </div>
  );
}
