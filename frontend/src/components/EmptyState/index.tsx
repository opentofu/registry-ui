import { Icon } from "@/components/Icon";
import { Paragraph } from "@/components/Paragraph";
import { empty } from "@/icons/empty";
import clsx from "clsx";

interface EmptyStateProps {
  text: string;
  className?: string;
}

export function EmptyState({ text, className }: EmptyStateProps) {
  return (
    <Paragraph className={clsx("flex flex-col items-center gap-2", className)}>
      <Icon path={empty} className="size-7" />
      {text}
    </Paragraph>
  );
}
