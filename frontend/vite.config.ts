import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import { Plugin, unified } from "unified";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeStringify from "rehype-stringify";
import remarkFrontmatter from "remark-frontmatter";
import { matter } from "vfile-matter";

const extractFrontmatter: Plugin = () => {
  return (_, file) => {
    matter(file);
  };
};

export default defineConfig({
  plugins: [
    react(),
    tsconfigPaths(),
    {
      name: "md-to-html",
      async transform(src, id) {
        if (id.endsWith(".md")) {
          const content = unified()
            .use(remarkParse)
            .use(remarkFrontmatter)
            .use(extractFrontmatter)
            .use(remarkRehype)
            .use(rehypeStringify)
            .processSync(src);

          return {
            code: `
              export const content = ${JSON.stringify(content.value)};
              export const frontmatter = ${JSON.stringify(content.data.matter)};
            `,
            map: null,
          };
        }
      },
    },
  ],
});
