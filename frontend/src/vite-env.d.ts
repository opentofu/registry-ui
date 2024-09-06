/// <reference types="vite/client" />

declare module "*.md" {
  export const content: string;
  export const frontmatter: Record<string, unknown>;
}
