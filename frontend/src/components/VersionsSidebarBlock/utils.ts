interface GroupVersionsOptions {
  versions: Array<{
    id: string;
    published: string;
  }>;
  currentVersion: string;
  latestVersion: string;
  versionLink: (version: string) => string;
}

interface VersionChild {
  label: string;
  published: string;
  isActive: boolean;
  id: string;
  link: string;
}

interface VersionGroup {
  label: string;
  published: string;
  children: VersionChild[];
  isActive: boolean;
  link: string;
}

interface GroupedVersionsAccumulator {
  [major: string]: VersionGroup;
}

export function groupVersions({
  versions,
  currentVersion,
  latestVersion,
  versionLink,
}: GroupVersionsOptions): VersionGroup[] {
  const groupedVersions = versions.reduce<GroupedVersionsAccumulator>((acc, version) => {
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
        link: "", // Version groups don't have individual links
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
  }, {});

  return Object.values(groupedVersions);
}
