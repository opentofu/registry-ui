import { content } from "./content" with { type: "macro" };

export function AnnouncementBar() {
  return (
    <div
      className="relative top-1.5 flex min-h-12 items-center justify-center font-medium"
      role="banner"
    >
      <a 
        href="#" 
        className="group block w-full no-underline transition-all duration-200 font-semibold bg-black/5 hover:bg-black/10 dark:bg-white/5 dark:hover:bg-white/15 text-gray-900 dark:text-white"
      >
        <div className="flex items-center justify-center gap-2 text-[15px] py-1.5">
          <span dangerouslySetInnerHTML={{ __html: content }} />
          <span className="inline-flex items-center font-bold ml-1 transition-transform duration-200 group-hover:translate-x-1">→</span>
        </div>
      </a>
    </div>
  );
}
