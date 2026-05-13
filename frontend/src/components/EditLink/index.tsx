interface EditLinkProps {
  url: string;
}

export function EditLink({ url }: EditLinkProps) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="text-brand-700 dark:text-brand-600 mt-5 inline-flex hover:underline"
    >
      Edit this page
    </a>
  );
}
