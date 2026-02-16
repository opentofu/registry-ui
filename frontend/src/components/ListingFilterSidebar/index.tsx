interface ListingFilterSidebarProps {
  sortBy: "alphabetical" | "popularity" | "recent";
  onSortChange: (sort: "alphabetical" | "popularity" | "recent") => void;
}

export function ListingFilterSidebar({
  sortBy,
  onSortChange,
}: ListingFilterSidebarProps) {
  return (
    <div className="p-4 space-y-6">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">
          Sort by
        </h3>
        <div className="space-y-2">
          <label className="flex items-center">
            <input
              type="radio"
              name="sort"
              value="alphabetical"
              checked={sortBy === "alphabetical"}
              onChange={(e) => onSortChange(e.target.value as any)}
              className="mr-2 text-brand-600 focus:ring-brand-500"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Alphabetical
            </span>
          </label>
          <label className="flex items-center">
            <input
              type="radio"
              name="sort"
              value="popularity"
              checked={sortBy === "popularity"}
              onChange={(e) => onSortChange(e.target.value as any)}
              className="mr-2 text-brand-600 focus:ring-brand-500"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Most Popular
            </span>
          </label>
          <label className="flex items-center">
            <input
              type="radio"
              name="sort"
              value="recent"
              checked={sortBy === "recent"}
              onChange={(e) => onSortChange(e.target.value as any)}
              className="mr-2 text-brand-600 focus:ring-brand-500"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Recently Updated
            </span>
          </label>
        </div>
      </div>
    </div>
  );
}
