import { fileURLToPath, URL } from 'url';

import { defineConfig } from 'vite';
import Vue from '@vitejs/plugin-vue';
import Icons from 'unplugin-icons/vite'
import IconsResolver from 'unplugin-icons/resolver'
import Components from 'unplugin-vue-components/vite'

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
    plugins: [
        Vue(),
        Components({
            resolvers: [
              IconsResolver({
                prefix: 'icon',
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
