interface GroupVersionsOptions {
  versions: Array<{
    id: string;
    published: string;
  }>;
  currentVersion: string;
  latestVersion: string;
  versionLink: (version: string) => string;
}

export function groupVersions({
  versions,
  currentVersion,
  latestVersion,
  versionLink,
}: GroupVersionsOptions) {
  const groupedVersions = versions.reduce((acc, version) => {
    if (version.id === latestVersion) {
      return acc;
    }

    const [major] = version.id.split(".");

    if (!acc[major]) {
      acc[major] = {
        label: `${major}.x`,
        published: version.published,
        children: [],
        isActive: false,
      };
    }

    const isActive = version.id === currentVersion;

    if (isActive) {
      acc[major].isActive = true;
    }

    acc[major].children.push({
      label: version.id,
      published: version.published,
      isActive,
      id: version.id,
      link: versionLink(version.id),
    });

    return acc;
  }, []);

  return Object.values(groupedVersions);
}
