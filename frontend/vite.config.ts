import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import macros from "unplugin-macros/vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [macros(), react(), tsconfigPaths(), tailwindcss()],
});
