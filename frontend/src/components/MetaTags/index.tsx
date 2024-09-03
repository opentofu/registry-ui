import { Helmet } from "react-helmet-async";

interface MetaTagsProps {
  title: string;
  description?: string;
}

export function MetaTags({ title, description }: MetaTagsProps) {
  const siteTitle = title
    ? `${title} - OpenTofu Registry`
    : "OpenTofu Registry";

  return (
    <Helmet>
      <title>{siteTitle}</title>
      <meta property="og:title" content={title} />
      <meta name="twitter:title" content={title} />
      {description && <meta name="description" content={description} />}
      {description && <meta property="og:description" content={description} />}
      {description && <meta name="twitter:description" content={description} />}
      <meta property="og:url" content={location.href} />
    </Helmet>
  );
}
