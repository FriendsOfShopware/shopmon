import { URL, fileURLToPath } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import Vue from "@vitejs/plugin-vue";
import IconsResolver from "unplugin-icons/resolver";
import Icons from "unplugin-icons/vite";
import Components from "unplugin-vue-components/vite";
import { defineConfig } from "vite";

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
