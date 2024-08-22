interface EditLinkProps {
  url: string;
}

export function EditLink({ url }: EditLinkProps) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="mt-5 inline-flex text-brand-700 hover:underline dark:text-brand-600"
    >
      Edit this page
    </a>
  );
}
