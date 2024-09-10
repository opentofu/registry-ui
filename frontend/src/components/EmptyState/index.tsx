import { Icon } from "@/components/Icon";
import { Paragraph } from "@/components/Paragraph";
import { search } from "@/icons/search";
import clsx from "clsx";

interface EmptyStateProps {
  text: string;
  className?: string;
}

export function EmptyState({ text, className }: EmptyStateProps) {
  return (
    <Paragraph className={clsx("flex items-center gap-2", className)}>
      <Icon path={search} className="size-4" />
      {text}
    </Paragraph>
  );
}
