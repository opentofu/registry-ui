import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import { unified } from "unified";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeStringify from "rehype-stringify";

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
            .use(remarkRehype)
            .use(rehypeStringify)
            .processSync(src);

          return {
            code: `export default ${JSON.stringify(content.value)}`,
            map: null,
          };
        }
      },
    },
  ],
});
