import { LoaderFunction } from "react-router";

import sidebar from "../../../docs/sidebar.json";

export type DocsRouteContext = {
  section?: string;
  subsection?: string;

  sectionBreadcrumbLabel?: string;
  subsectionBreadcrumbLabel?: string;
};

const getBreadcrumbLabel = (section?: string, subsection?: string): string | undefined => {
  // find the section first
  if (!section) {
    return "Docs";
  }

  const sectionItem = sidebar.find((item) => item.slug === section);
  if (!sectionItem) {
    return section;
  }
  
  if (!subsection) {
    return sectionItem.title || section;
  }
  const slug = `${section}/${subsection}`;
  const subsectionItem = sectionItem.items?.find((item) => item.slug === slug);
  if (subsectionItem) {
    return subsectionItem.title || subsection;
  }
  return undefined
};

export const docsMiddleware: LoaderFunction = async ({ params }, context) => {
  const { section, subsection } = params;

  // Attach the section and subsection to the context
  const docsContext = context as DocsRouteContext;
  docsContext.section = section;
  docsContext.subsection = subsection;

  docsContext.sectionBreadcrumbLabel = getBreadcrumbLabel(section) || 'BROKEN LINK';
  docsContext.subsectionBreadcrumbLabel = getBreadcrumbLabel(section, subsection) || 'BROKEN LINK';
}