/// <reference types="vite/client" />
/// <reference types="unplugin-icons/types/vue3" />

declare module "*.vue" {
  import type { DefineComponent } from "vue";

  const component: DefineComponent<Record<string, never>, Record<string, never>, unknown>;
  export default component;
}
