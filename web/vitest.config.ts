import path from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig, type Plugin } from "vitest/config";

function stubCss(): Plugin {
  return {
    name: "stub-css",
    transform(_code, id) {
      if (id.endsWith(".css")) {
        return { code: "export default {}", map: null };
      }
    },
  };
}

export default defineConfig({
  plugins: [stubCss(), react()],
  test: {
    include: ["tests/**/*.test.{ts,tsx}"],
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
    globals: true,
    server: {
      deps: {
        inline: true,
      },
    },
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "tests/**",
        "src/app/layout.tsx",
        "src/app/**/layout.tsx",
      ],
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@test": path.resolve(__dirname, "./tests"),
    },
  },
});
