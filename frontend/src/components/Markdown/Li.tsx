import { HTMLAttributes } from "react";

export function MarkdownLi({ children, id }: HTMLAttributes<HTMLLIElement>) {
  return (
    <li 
      id={id}
      className="pl-2 leading-7 text-gray-700 dark:text-gray-300 [&+&]:mt-2.5"
    >
      {children}
    </li>
  );
}
