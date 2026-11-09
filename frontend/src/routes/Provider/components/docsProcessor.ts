// Markdown documentation for terraform providers often contains relative paths to other docs within that provider/module
// we want to handle this situation, but our pathing is slightly different to the terraform registry
// so we need to rework the relative paths to point to the correct location
/*
  Cases:
  stripping extensions:
  - /docs/overview.md -> /docs/overview
  - /docs/overview.html -> /docs/overview

  anchor links
  - /docs/overview.md#section -> /docs/overview#section
  - /docs/overview.html#section -> /docs/overview#section

  directory structure changes:ro
  - docs/providers/<providername> -> /provider/<namespace>/<providername>/<version>/docs/
  - docs/providers/<providername>/r/<resource> -> /provider/<namespace>/<providername>/<version>/docs/resources/<resource>
  - docs/providers/<providername>/d/<data> -> /provider/<namespace>/<providername>/<version>/docs/datasources/<data>
*/

type DocProcessor = (
  path: string,
  namespace?: string,
  provider?: string,
  version?: string,
) => string;

export const extensionsToStrip = [
  ".html.markdown",
  ".html.md",
  ".html",
  ".markdown.html",
  ".markdown",
  ".md.html",
  ".md",
];

// stripExtension is used to remove extensions from a path
// but keep the anchor link if it exists
export const stripExtension: DocProcessor = (path: string): string => {
  return extensionsToStrip.reduce((acc, ext) => {
    const regex = new RegExp(`(${ext})(#.*)?$`);
    return acc.replace(regex, "$2");
  }, path);
};

// shortToLongPath is used to convert a short path to a long path
// for example: /r/ -> /resources/
export const shortToLongPath = (path: string): string => {
  const shortToLongMap: Record<string, string> = {
    r: "resources",
    d: "datasources",
    "data-sources": "datasources",
    f: "functions",
  };

  const splitPath = path.split("/");
  for (let i = 0; i < splitPath.length; i++) {
    const part = splitPath[i];
    if (shortToLongMap[part]) {
      splitPath[i] = shortToLongMap[part];
    }
  }

  return splitPath.join("/");
};

// Helper to process a single path through all processors
const processPath = (
  path: string,
  namespace?: string,
  provider?: string,
  version?: string,
): string => {
  const processors: Array<DocProcessor> = [stripExtension, shortToLongPath];

  return processors.reduce((acc, processor) => {
    return processor(acc, namespace, provider, version);
  }, path);
};

// Helper to add user-content- prefix to anchors in paths as rehypeSanitize has
// added this prefix to all anchors in the output to avoid clobbering attacks
export const prefixAnchors = (path: string): string => {
  const anchorRegex = /#(.*)/;
  const match = path.match(anchorRegex);
  if (match) {
    const anchor = match[1];
    const newPath = path.replace(anchorRegex, "");
    return `${newPath}#user-content-${anchor}`;
  }
  return path;
};

export const reworkRelativePaths = (
  doc: string,
  namespace: string,
  provider: string,
  version: string,
  isOverview = false,
): string => {
  // Long links!
  const linkRegex = /(\[.*?\]\()([^)]+)(\))/g;

  // Iterate across all links
  const reworked = doc.replace(linkRegex, (match, prefix, path, suffix) => {
    path = prefixAnchors(path);

    if (path.startsWith("#")) {
      return `${prefix}${path}${suffix}`;
    }

    const docsPath = path.startsWith("/docs/providers/");
    const relativePath =
      !path.startsWith("/") && !/^[a-z][a-z\d+.-]*:/i.test(path);

    if (relativePath) {
      const processedPath = processPath(path);
      if (!isOverview) {
        return `${prefix}${processedPath}${suffix}`;
      }

      const docsRelativePath = processedPath.replace(/^\.\//, "");
      return `${prefix}/provider/${namespace}/${provider}/${version}/docs/${docsRelativePath}${suffix}`;
    }

    // Otherwise, we only care about links to /docs/providers.
    if (!docsPath) {
      return match;
    }

    // Extract the provider names from the path.
    const providerRegex = /\/docs\/providers\/([^/]+)/;
    const providerMatch = path.match(providerRegex);
    if (providerMatch) {
      // Only transform links for the current provider.
      if (providerMatch[1] !== provider) {
        return match;
      }
    }

    const parts = path.split("/");
    if (parts.length >= 4) {
      const remainingPath = parts.slice(4).join("/");
      const processedPath = processPath(
        remainingPath,
        namespace,
        provider,
        version,
      );

      return `${prefix}/provider/${namespace}/${provider}/${version}/docs/${processedPath}${suffix}`;
    }

    return match;
  });

  return reworked;
};
