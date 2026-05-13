import { content } from "./content" with { type: "macro" };

export function AnnouncementBar() {
  return (
    <div
      className="bg-brand-600 flex h-6 items-center justify-center text-sm text-gray-900 [&_a]:underline"
      role="banner"
      dangerouslySetInnerHTML={{ __html: content }}
    />
  );
}
