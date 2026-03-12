interface HeadingLinkProps {
  id: string;
  label: string;
}

export function HeadingLink({ id, label }: HeadingLinkProps) {
  return (
    <a
      href={`#${id}`}
      aria-label={`Direct link to ${label}`}
      className="text-brand-600 ml-2 hidden group-hover:inline hover:underline hover:underline-offset-2"
    >
      #
    </a>
  );
}
