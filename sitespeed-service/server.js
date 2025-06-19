const express = require('express');
const { exec } = require('node:child_process');
const fs = require('node:fs').promises;
const path = require('node:path');
const { argv } = require('node:process');

const app = express();
const PORT = 3001;

app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
    res.json({ status: 'ok', service: 'sitespeed-service' });
});

// Main endpoint to run sitespeed analysis
app.post('/analyze', async (req, res) => {
    const { shopId, urls, label, folderName } = req.body;

    if (!shopId || !urls || !Array.isArray(urls) || urls.length === 0) {
        return res.status(400).json({
            error: 'Missing required parameters: shopId and urls',
        });
    }

    try {
        // Create results directory for this shop and URL
        const dataFolder =
            process.env.APP_SITESPEED_DATA_FOLDER || '/app/results';

        const resultsDir = path.join(dataFolder, shopId.toString());
        await fs.mkdir(resultsDir, { recursive: true });

        formattedUrls = urls.map((url) => {
            url.replace(
                'http://localhost:3889',
                'http://shopmon-demoshop-1:8000',
            );
        });

        // Run sitespeed.io analysis with all metrics including visual metrics and transfer size
        // Note: Cannot use --headless with --video/--visualMetrics as they need a screen
        const command = `/usr/src/app/bin/sitespeed.js --outputFolder ${resultsDir} --plugins.add analysisstorer --visualMetrics --video --browsertime.iterations 1 "${formattedUrls.join(' ')}"`;

        console.log(
            `Running sitespeed analysis for shop ${shopId}`,
        );

        exec(command, async (error, stdout, stderr) => {
            console.log('Command stdout:', stdout);
            console.log('Command stderr:', stderr);

            if (error) {
                console.error(`Error running sitespeed: ${error}`);
                return res.status(500).json({
                    error: 'Failed to run sitespeed analysis',
                    details: error.message,
                    stdout: stdout,
                    stderr: stderr,
                });
            }

            try {
                // List files in results directory to see what was created
                const files = await fs.readdir(resultsDir);
                console.log('Files in results directory:', files);

                // Look for data folder with JSON files
                const dataDir = path.join(resultsDir, 'data');
                try {
                    const dataFiles = await fs.readdir(dataDir);
                    console.log('Files in data directory:', dataFiles);

                    // Look for browsertime summary files
                    let browsertimeData;
                    let pagexrayData;
                    const browsertimeFiles = dataFiles.filter(
                        (f) =>
                            f.startsWith('browsertime.summary-') &&
                            f.endsWith('.json'),
                    );
                    const pagexrayFiles = dataFiles.filter(
                        (f) =>
                            f.startsWith('pagexray.summary-') &&
                            f.endsWith('.json') &&
                            !f.includes('total'),
                    );

                    if (browsertimeFiles.length > 0) {
                        const browsertimeFile = path.join(
                            dataDir,
                            browsertimeFiles[0],
                        );
                        const browsertimeContent = await fs.readFile(
                            browsertimeFile,
                            'utf8',
                        );
                        browsertimeData = JSON.parse(browsertimeContent);
                        console.log(`Using ${browsertimeFiles[0]}`);
                    } else {
                        throw new Error(
                            'No browsertime summary JSON file found in data directory',
                        );
                    }

                    // Also load pagexray data for transfer size
                    if (pagexrayFiles.length > 0) {
                        const pagexrayFile = path.join(
                            dataDir,
                            pagexrayFiles[0],
                        );
                        const pagexrayContent = await fs.readFile(
                            pagexrayFile,
                            'utf8',
                        );
                        pagexrayData = JSON.parse(pagexrayContent);
                        console.log(
                            `Using ${pagexrayFiles[0]} for additional metrics`,
                        );
                    }

                    const metrics = extractMetrics(
                        browsertimeData,
                        pagexrayData,
                    );

                    console.log(
                        `Analysis completed for shop ${shopId} - ${label || 'Unlabeled'} (${safeName})`,
                        metrics,
                    );
                    res.json({
                        shopId,
                        url,
                        label: label || null,
                        folderName: safeName,
                        timestamp: new Date().toISOString(),
                        metrics,
                    });
                } catch (fileError) {
                    console.error(`Data file error: ${fileError}`);
                    res.status(500).json({
                        error: 'Data files not found or invalid',
                        details: fileError.message,
                        files: files,
                    });
                }
            } catch (parseError) {
                console.error(`Error parsing results: ${parseError}`);
                res.status(500).json({
                    error: 'Failed to parse sitespeed results',
                    details: parseError.message,
                });
            }
        });
    } catch (error) {
        console.error(`Setup error: ${error}`);
        res.status(500).json({
            error: 'Failed to setup analysis',
            details: error.message,
        });
    }
});

function extractMetrics(browsertimeData, pagexrayData = null) {
    const metrics = {};

    // Time to First Byte (TTFB)
    if (browsertimeData.googleWebVitals?.ttfb) {
        metrics.ttfb = Math.round(
            browsertimeData.googleWebVitals.ttfb.median || 0,
        );
    }

    // Fully Loaded Time
    if (browsertimeData.timings?.fullyLoaded) {
        metrics.fullyLoaded = Math.round(
            browsertimeData.timings.fullyLoaded.median || 0,
        );
    }

    // Largest Contentful Paint
    if (browsertimeData.googleWebVitals?.largestContentfulPaint) {
        metrics.largestContentfulPaint = Math.round(
            browsertimeData.googleWebVitals.largestContentfulPaint.median || 0,
        );
    }

    // First Contentful Paint
    if (browsertimeData.googleWebVitals?.firstContentfulPaint) {
        metrics.firstContentfulPaint = Math.round(
            browsertimeData.googleWebVitals.firstContentfulPaint.median || 0,
        );
    }

    // Cumulative Layout Shift
    if (browsertimeData.googleWebVitals?.cumulativeLayoutShift) {
        metrics.cumulativeLayoutShift = Number.parseFloat(
            (
                browsertimeData.googleWebVitals.cumulativeLayoutShift.median ||
                0
            ).toFixed(3),
        );
    }

    // Speed Index
    if (browsertimeData.visualMetrics?.SpeedIndex) {
        metrics.speedIndex = Math.round(
            browsertimeData.visualMetrics.SpeedIndex.median || 0,
        );
    }

    // Page Load Time
    if (browsertimeData.pageTimings?.pageLoadTime) {
        metrics.pageLoadTime = Math.round(
            browsertimeData.pageTimings.pageLoadTime.median || 0,
        );
    }

    // Total Blocking Time
    if (browsertimeData.googleWebVitals?.totalBlockingTime) {
        metrics.totalBlockingTime = Math.round(
            browsertimeData.googleWebVitals.totalBlockingTime.median || 0,
        );
    }

    // Transfer Size (from pagexray data)
    if (pagexrayData?.transferSize?.median) {
        metrics.transferSize = Math.round(pagexrayData.transferSize.median);
    } else if (
        browsertimeData.transferSize &&
        typeof browsertimeData.transferSize === 'number'
    ) {
        metrics.transferSize = Math.round(browsertimeData.transferSize);
    } else if (browsertimeData.statistics?.transferSize?.median) {
        metrics.transferSize = Math.round(
            browsertimeData.statistics.transferSize.median,
        );
    }

    return metrics;
}

app.listen(PORT, '0.0.0.0', () => {
    console.log(`Sitespeed service running on port ${PORT}`);
});
