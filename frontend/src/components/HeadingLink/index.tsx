interface HeadingLinkProps {
  id: string;
  label: string;
}

export function HeadingLink({ id, label }: HeadingLinkProps) {
  return (
    <a
      href={`#${id}`}
      aria-label={`Direct link to ${label}`}
      className="ml-2 hidden text-brand-600 hover:underline hover:underline-offset-2 group-hover:inline"
    >
      #
    </a>
  );
}
