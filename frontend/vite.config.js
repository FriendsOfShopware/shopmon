import { fileURLToPath, URL } from 'url';

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const hmr = {};

if (process.env.HMR_PORT) {
    hmr.clientPort = process.env.HMR_PORT;
}

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url))
        }
    },
    server: {
        hmr: {
            ...hmr
        },
        proxy: {
            '/api': {
                target: process.env.SHOPMON_API_URL || 'https://shopmon.fos.gg',
                changeOrigin: true
            }
        }
    }
});
