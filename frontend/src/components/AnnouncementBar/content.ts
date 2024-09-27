import { unified } from "unified";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import rehypeStringify from "rehype-stringify";
import announcement from "../../../announcement.md?raw";

export const content = unified()
  .use(remarkParse)
  .use(remarkRehype)
  .use(rehypeStringify)
  .processSync(announcement).value;
