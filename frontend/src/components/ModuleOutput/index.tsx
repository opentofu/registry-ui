import { HeadingLink } from "../HeadingLink";
import { Paragraph } from "../Paragraph";

interface ModuleOutputProps {
  name: string;
  description: string;
}

export function ModuleOutput({ name, description }: ModuleOutputProps) {
  return (
    <li>
      <h4 id={name} className="group scroll-mt-5">
        {name}
        <HeadingLink id={name} label={`${name} output`} />
      </h4>
      <Paragraph className="mt-1">{description}</Paragraph>
    </li>
  );
}
