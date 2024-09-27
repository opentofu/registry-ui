import { content } from "./content" with { type: "macro" };

export function AnnouncementBar() {
  return (
    <div
      className="flex h-6 items-center justify-center bg-brand-600 text-sm text-gray-900 [&_a]:underline"
      role="banner"
      dangerouslySetInnerHTML={{ __html: content }}
    />
  );
}
