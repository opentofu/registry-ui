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

  directory structure changes:
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
): string => {
  // Long links!
  const linkRegex = /(\[.*?\]\()([^)]+)(\))/g;

  // Iterate across all links
  const reworked = doc.replace(linkRegex, (match, prefix, path, suffix) => {
    path = prefixAnchors(path);

    // its just a simple file link, let's not mess with it but just strip the extension
    if (!path.includes("/")) {
      const strippedPath = stripExtension(path);
      return `${prefix}${strippedPath}${suffix}`;
    }

    // otherwise, we only care about links to /docs/providers
    if (!path.startsWith("/docs/providers/")) {
      // If the path doesn't start with /docs/providers/, return the original link
      return match;
    }

    // extract the provider names from the path
    const providerRegex = /\/docs\/providers\/([^/]+)/;
    const providerMatch = path.match(providerRegex);
    if (providerMatch) {
      // Only transform links for the current provider,
      // if the path is for a different provider, return the original link
      // its best to not mess with it for now
      if (providerMatch[1] !== provider) {
        return match;
      }
    }

    const parts = path.split("/");
    if (parts.length >= 4) {
      const remainingPath = parts.slice(4).join("/");
      const processedPath = processPath(
        `${remainingPath}`,
        namespace,
        provider,
        version,
      );

      // Transform to registry path structure
      return `${prefix}/provider/${namespace}/${provider}/${version}/docs/${processedPath}${suffix}`;
    }

    // Return unmodified if it doesn't match our patterns
    return prefix + path + suffix;
  });

  return reworked;
};
