/// <reference types="vite/client" />
/// <reference types="unplugin-icons/types/vue3" />

declare module '*.vue' {
    import type { DefineComponent } from 'vue'
    // eslint-disable-line
    const component: DefineComponent<{}, {}, any>
    export default component
}
  