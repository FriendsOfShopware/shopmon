import { ViteSSG } from "vite-ssg";

import "./style.css";

import App from "./App.vue";
import { routes, setupRouterGuards } from "./router";
import { i18n } from "./i18n";

export const createApp = ViteSSG(
  App,
  {
    routes,
    linkActiveClass: "active",
  },
  ({ app, router, isClient }) => {
    app.use(i18n as any);

    if (isClient) {
      setupRouterGuards(router);
    }
  },
);
