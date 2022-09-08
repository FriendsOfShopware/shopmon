import { fileURLToPath, URL } from 'url';

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const hmr = {};

if (process.env.HMR_PORT) {
    hmr.clientPort = process.env.HMR_PORT;
}

let wssServer = 'wss://shopmon.fos.gg';

if (process.env.SHOPMON_API_URL) {
    const url = new URL(process.env.SHOPMON_API_URL);

    if (url.protocol === 'https:') {
        url.protocol = 'wss:';
    } else {
        url.protocol = 'ws:';
    }

    wssServer = url.toString();
}

export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url))
        }
    },
    server: {
        hmr,
        proxy: {
            '/api/ws': {
                target: wssServer,
                ws: true,
                changeOrigin: true
            },
            '/api': {
                target: process.env.SHOPMON_API_URL || 'https://shopmon.fos.gg',
                changeOrigin: true
            }
        }
    }
});
