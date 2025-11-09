import {Hono} from 'hono';
import {serve} from '@hono/node-server';
import { exec } from 'node:child_process';
import { existsSync, promises as fs, readdirSync } from 'node:fs';
import path from 'node:path';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';
import { promisify } from 'node:util';
import os from 'node:os';

const dataFolder = process.env.APP_SITESPEED_DATA_FOLDER || '/app/results';

if (!existsSync(dataFolder)) {
    await fs.mkdir(dataFolder, { recursive: true });
}

/**
 * Clean up Chromium temporary directories in /tmp that are older than the specified age
 * These are created during sitespeed analysis and need to be removed
 * @param maxAgeMinutes - Only delete directories older than this many minutes (default: 5)
 */
async function cleanupChromiumTempFiles(maxAgeMinutes = 5) {
    const tmpDir = os.tmpdir();
    const maxAgeMs = maxAgeMinutes * 60 * 1000;
    const now = Date.now();

    try {
        const files = await fs.readdir(tmpDir);
        const chromiumDirs = files.filter(file => file.startsWith('.org.chromium.Chromium.'));

        for (const dir of chromiumDirs) {
            const fullPath = path.join(tmpDir, dir);
            try {
                const stat = await fs.stat(fullPath);
                if (stat.isDirectory()) {
                    const ageMs = now - stat.mtimeMs;
                    if (ageMs > maxAgeMs) {
                        await fs.rm(fullPath, { recursive: true, force: true });
                        console.log(`Cleaned up Chromium temp directory (${Math.round(ageMs / 60000)}min old): ${fullPath}`);
                    }
                }
            } catch (err) {
                console.error(`Failed to clean up ${fullPath}:`, err);
            }
        }
    } catch (err) {
        console.error('Failed to clean up Chromium temp files:', err);
    }
}

const app = new Hono();

app.get('/health', (c) => {
    return c.json({ status: 'ok', service: 'sitespeed-service' });
});

type BrowserTime = {
    timings: {
        fullyLoaded: { median: number };
        largestContentfulPaint: { median: number };
    }
    googleWebVitals: {
        ttfb: { median: number };
        largestContentfulPaint: { median: number };
        firstContentfulPaint: { median: number };
        cumulativeLayoutShift: { median: number };
        totalBlockingTime: { median: number };
    };
}

type PageXray = {
    transferSize: { median: number };
}

app.post('/analyze',
 zValidator('json', z.object({
     shopId: z.number(),
     urls: z.array(z.string().url()).min(1).max(5),
 })), async (c) => {
    const validated = c.req.valid('json');
    const { shopId, urls } = validated;

    const resultDir = path.join(dataFolder, shopId.toString());
    if (!existsSync(resultDir)) {
        await fs.mkdir(resultDir, { recursive: true });
    } else {
        await fs.rmdir(resultDir, { recursive: true });
        await fs.mkdir(resultDir, { recursive: true });
    }

    const formattedUrls = urls.map((url) => {
        return url.replace('http://localhost:3889', 'http://demoshop:8000');
    });

    console.log(`Starting sitespeed analysis for shop ${shopId} with URLs: ${formattedUrls.join(', ')}`);

    const command = `/usr/src/app/bin/sitespeed.js --outputFolder ${resultDir} --plugins.add analysisstorer --visualMetrics --video --viewPort 1920x1080 --browsertime.chrome.cleanUserDataDir=true --browsertime.iterations 1 "${formattedUrls.join('" "')}"`;
    console.log(`Running sitespeed analysis for shop ${shopId} with command: ${command}`);

    try {
        const { stdout, stderr } = await promisify(exec)(command);
        console.log(`Sitespeed analysis completed for shop ${shopId}`);

        const pages = readdirSync(path.join(resultDir, 'pages'));

        const webvitalDataPath = path.join(resultDir, 'data/browsertime.summary-total.json')
        const pagexrayDataPath = path.join(resultDir, 'data/pagexray.summary-total.json');

        if (!existsSync(webvitalDataPath)) {
            await fs.rmdir(resultDir, { recursive: true });
            return c.json({
                error: 'Web vital data not found',
            });
        }

        const browsertimeData = JSON.parse(
            await fs.readFile(webvitalDataPath, 'utf-8'),
        ) as BrowserTime;

        const pagexrayData = JSON.parse(
            existsSync(pagexrayDataPath)
                ? await fs.readFile(pagexrayDataPath, 'utf-8')
                : '{}',
        ) as PageXray;

        return c.json({
            ttfb: browsertimeData.googleWebVitals?.ttfb?.median || 0,
            fullyLoaded: browsertimeData.timings?.fullyLoaded?.median || 0,
            largestContentfulPaint: browsertimeData.googleWebVitals?.largestContentfulPaint?.median || 0,
            firstContentfulPaint: browsertimeData.googleWebVitals?.firstContentfulPaint?.median || 0,
            cumulativeLayoutShift: browsertimeData.googleWebVitals?.cumulativeLayoutShift?.median || 0,
            transferSize: pagexrayData.transferSize?.median || 0,
            screenshotPath: path.join('pages', pages[0], 'data', 'screenshots', '1', 'afterPageCompleteCheck.png')
        });
    } catch (error) {
        console.error(`Error running sitespeed: ${error}`);
        return c.json({
            error: 'Failed to run sitespeed analysis',
            details: error.message,
        }, 500);
    } finally {
        // Clean up Chromium temp files after each run
        await cleanupChromiumTempFiles();
    }
});

// Run cleanup every 5 minutes for files older than 5 minutes
const CLEANUP_INTERVAL_MS = 5 * 60 * 1000; // 5 minutes
setInterval(() => {
    cleanupChromiumTempFiles(5);
}, CLEANUP_INTERVAL_MS);

// Run initial cleanup on startup
cleanupChromiumTempFiles(5);

serve({
  fetch: app.fetch,
  port: 3001
}, (info) => {
    console.log(`Sitespeed service is running on http://0.0.0.0:${info.port}`);
    console.log('Chromium temp file cleanup scheduled every 5 minutes');
});
