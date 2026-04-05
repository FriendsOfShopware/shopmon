import { URL, fileURLToPath } from "node:url";
import { writeFileSync } from "node:fs";
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

const hmr = {};

if (process.env.HMR_PORT) {
  hmr.clientPort = process.env.HMR_PORT;
}

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
    hmr,
    proxy: {
      "/api": {
        target: process.env.SHOPMON_API_URL || "https://shopmon.fos.gg",
        changeOrigin: true,
      },
    },
  },
});
