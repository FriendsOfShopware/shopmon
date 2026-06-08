import { URL, fileURLToPath } from "node:url";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

import tailwindcss from "@tailwindcss/vite";
import Vue from "@vitejs/plugin-vue";
import IconsResolver from "unplugin-icons/resolver";
import Icons from "unplugin-icons/vite";
import Components from "unplugin-vue-components/vite";
import { defineConfig } from "vite";

const ssgRoutes = [
  "/",
  "/privacy",
  "/imprint",
  "/account/login",
  "/account/register",
  "/account/forgot-password",
];

export default defineConfig({
  plugins: [
    tailwindcss(),
    Vue(),
    Components({
      resolvers: [
        IconsResolver({
          prefix: "icon",
        }),
      ],
    }),
    Icons({
      autoInstall: true,
      scale: 1,
    }),
  ],
  ssgOptions: {
    includedRoutes() {
      return ssgRoutes;
    },
    dirStyle: "nested",
    onFinished() {
      const siteUrl = "https://shopmon.fos.gg";
      const today = new Date().toISOString().split("T")[0];
      const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${ssgRoutes.map((route) => `  <url>\n    <loc>${siteUrl}${route}</loc>\n    <lastmod>${today}</lastmod>\n  </url>`).join("\n")}
</urlset>`;
      writeFileSync(resolve("dist", "sitemap.xml"), sitemap);

      // Generate a clean SPA fallback for /app/* routes so users
      // don't briefly see the SSG'd landing page before Vue hydrates.
      const indexHtml = readFileSync(resolve("dist", "index.html"), "utf-8");
      const appDivStart = indexHtml.indexOf('<div id="app"');
      const afterAppDiv = indexHtml.indexOf("<link", appDivStart);
      const appFallback =
        indexHtml.substring(0, appDivStart) +
        '<div id="app" class="h-full"></div>\n' +
        indexHtml.substring(afterAppDiv);
      mkdirSync(resolve("dist", "app"), { recursive: true });
      writeFileSync(resolve("dist", "app", "index.html"), appFallback);
    },
  },
  ssr: {
    noExternal: [/vue-i18n/],
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: process.env.SHOPMON_API_URL || "https://shopmon.fos.gg",
        changeOrigin: true,
      },
    },
  },
});
