import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import macros from "unplugin-macros/vite";
import tailwindcss from "@tailwindcss/vite";
import svgr from "vite-plugin-svgr";

export default defineConfig({
  plugins: [macros(), react(), svgr(), tailwindcss()],
  resolve: {
    tsconfigPaths: true,
  },
});
