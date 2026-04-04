import { createApp } from "vue";

import "./app.css";

import App from "./App.vue";
import { router } from "./router";
import { i18n } from "./i18n";

const app = createApp(App);

app.use(i18n as any);
app.use(router);

app.mount("#app");
