interface ProviderDocsEditLinkProps {
  url: string;
}

export function ProviderDocsEditLink({ url }: ProviderDocsEditLinkProps) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="text-brand-700 hover:underline dark:text-brand-600"
    >
      Edit this page
    </a>
  );
}
