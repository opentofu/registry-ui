import { content } from "./content" with { type: "macro" };

export function AnnouncementBar() {
  return (
    <div
      className="relative top-1.5 flex min-h-12 items-center justify-center font-medium"
      role="banner"
    >
      <a
        href="#"
        className="group block w-full bg-black/5 font-semibold text-gray-900 no-underline transition-all duration-200 hover:bg-black/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/15"
      >
        <div className="flex items-center justify-center gap-2 py-1.5 text-[15px]">
          <span dangerouslySetInnerHTML={{ __html: content }} />
          <span className="ml-1 inline-flex items-center font-bold transition-transform duration-200 group-hover:translate-x-1">
            →
          </span>
        </div>
      </a>
    </div>
  );
}
