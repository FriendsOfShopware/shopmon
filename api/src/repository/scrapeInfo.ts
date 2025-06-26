import { mkdir, readFile, rm, stat, writeFile } from 'node:fs/promises';
import { promisify } from 'node:util';
import { gunzip, gzip } from 'node:zlib';
import type { ShopScrapeInfo } from '../types/index.ts';

const pGzip = promisify(gzip);
const pGunzip = promisify(gunzip);

const dataFolder = 'files/shops';

async function getPath(shopId: number) {
    const shopFolder = `${dataFolder}/${shopId}`;

    try {
        await stat(shopFolder);
    } catch (_e) {
        await mkdir(shopFolder, { recursive: true });
    }

    return `${shopFolder}/scrape-info.json.gz`;
}

export async function getShopScrapeInfo(
    shopId: number,
): Promise<ShopScrapeInfo | null> {
    const path = await getPath(shopId);

    try {
        const file = await readFile(path);

        const decompressed = await pGunzip(file);

        return JSON.parse(decompressed.toString());
    } catch (_e) {
        return null;
    }
}

export async function saveShopScrapeInfo(shopId: number, info: ShopScrapeInfo) {
    const path = await getPath(shopId);

    const compressed = await pGzip(Buffer.from(JSON.stringify(info)));

    await writeFile(path, compressed);
}

export async function deleteShopScrapeInfo(shopId: number) {
    const path = await getPath(shopId);

    try {
        await rm(path);
    } catch (_e) {
        // Ignore errors if file does not exist
    }
}
