import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "~wails": path.resolve(__dirname, "./wailsjs/go"),
      "~wails/runtime": path.resolve(__dirname, "./wailsjs/runtime/runtime.js"),
    },
  },
});
