import { definitions } from "@/api";

export type NestedItem = {
  name: string;
  label: string;
  items?: NestedItem[];
  type?: string;
  open: boolean;
  active?: boolean;
};

export type Docs = Omit<definitions["ProviderDocs"], "index">;
type Category = keyof Docs;

// categoryLabelMap is a mapping of the keys in the OriginalStructure to a human readable name
const categoryLabelMap: Map<Category, string> = new Map([
  ["resources", "Resources"],
  ["datasources", "Data Sources"],
  ["functions", "Functions"],
  ["guides", "Guides"],
]);

// transformStructure takes the data from the ProviderDocs and transforms it into a NestedItem array
// representing the structure of the sidebar
// The currentType and currentDoc are used to determine which item should be marked as active
// and which categories should be open
// The result is an array of NestedItems representing the structure of the sidebar
export function transformStructure(
  data: Docs,
  currentType?: string,
  currentDoc?: string,
): NestedItem[] {
  const knownCategories = Object.keys(data) as Array<Category>;
  const result = new Map<string, NestedItem>();

  for (const category of knownCategories) {
    const categoryItems = data[category];
    const categoryLabel = categoryLabelMap.get(category as Category);

    if (!categoryLabel) {
      // We don't have a label for this category, so we skip it;
      continue;
    }

    for (const item of categoryItems) {
      const { name, title, subcategory } = item;

      // Determine if the item should be marked as active
      const isActive =
        currentDoc !== undefined &&
        currentType !== undefined &&
        currentType === category &&
        currentDoc === name;

      if (!subcategory) {
        // If no subcategory, add directly under the category
        if (!result.has(categoryLabel)) {
          result.set(categoryLabel, {
            name: category,
            label: categoryLabel,
            items: [],
            open: false,
          });
        }

        result.get(categoryLabel)!.items!.push({
          name: name || "",
          label: title || "",
          type: category,
          active: isActive,
          open: false,
        });

        // If the item is active, ensure the category is open
        if (isActive) {
          result.get(categoryLabel)!.open = true;
        }
      } else {
        // If subcategory exists, ensure it is created and nested
        if (!result.has(subcategory)) {
          result.set(subcategory, {
            name: subcategory,
            label: subcategory,
            items: [],
            open: false,
          });
        }

        const subcategoryItem = result.get(subcategory)!;

        // Ensure the category container (e.g., "resources" or "datasources") is created under the subcategory
        let categoryContainer = subcategoryItem.items!.find(
          (subItem) => subItem.name === category,
        );

        if (!categoryContainer) {
          categoryContainer = {
            name: category,
            label: categoryLabel,
            items: [],
            type: category,
            open: false,
          };
          subcategoryItem.items!.push(categoryContainer);
        }

        // Add the item under its respective category container within the subcategory
        categoryContainer.items!.push({
          name: name || "",
          label: title || "",
          type: category,
          active: isActive,
          open: false,
        });

        // If the item is active, ensure the subcategory and its parent category are open
        if (isActive) {
          subcategoryItem.open = true;
          categoryContainer.open = true;
        }
      }
    }
  }

  return sortSidebar(Array.from(result.values()));
}

// recurses through all levels of the sidebar and sort the items based on their label,
// placing known categories first
function sortSidebar(sidebar: NestedItem[]): NestedItem[] {
  const categories = Array.from(categoryLabelMap.values());
  sidebar = sidebar.sort((a, b) => {
    const aIndex = categories.indexOf(a.label);
    const bIndex = categories.indexOf(b.label);

    if (aIndex !== -1 && bIndex !== -1) {
      return aIndex - bIndex;
    }

    if (aIndex !== -1 && bIndex === -1) {
      return -1;
    }

    if (aIndex === -1 && bIndex !== -1) {
      return 1;
    }

    return a.label.localeCompare(b.label);
  });

  sidebar.forEach((item) => {
    if (item.items) {
      item.items = sortSidebar(item.items);
    }
  });

  return sidebar;
}
