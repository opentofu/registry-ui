interface MetaTagsProps {
  title?: string;
  description?: string;
}

export function MetaTags({ title, description }: MetaTagsProps) {
  const siteTitle = title
    ? `${title} - OpenTofu Registry`
    : "OpenTofu Registry";

  return (
    <>
      <title>{siteTitle}</title>
      <meta property="og:title" content={siteTitle} />
      <meta name="twitter:title" content={siteTitle} />
      {description && <meta name="description" content={description} />}
      {description && <meta property="og:description" content={description} />}
      {description && <meta name="twitter:description" content={description} />}
      <meta property="og:url" content={location.href} />
    </>
  );
}
